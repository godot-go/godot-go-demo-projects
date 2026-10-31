package demo

// Watchdog: dump goroutine stacks and kernel thread state when the game loop
// stops advancing, or on demand via SIGUSR1.
//
// Why this exists: Godot loads the extension as a c-shared library. In that
// host the usual Ctrl-\ (SIGQUIT) goroutine dump is unavailable — the Go
// runtime declines to install a SIGQUIT handler when the disposition is
// SIG_IGN at init (shells ignore SIGQUIT for background jobs), and the engine
// restores the default afterwards. A wedged game then dies without stacks:
// `kill -QUIT` produces a core file and nothing else.
//
// The watchdog runs on its own goroutine (pinned to an OS thread) and writes
// straight to a file, so it does not depend on the logger, the terminal, or
// the Godot main thread. It catches the common failure mode — a deadlock in a
// callback while other runtime threads keep running — and gives the kernel
// side (which thread is stuck in which wait channel) alongside the Go side.
//
// Environment:
//
//	GODOT_GO_WATCHDOG_FILE           dump path (default: $TMPDIR/godot-go-watchdog-<pid>.log)
//	GODOT_GO_WATCHDOG_STALL_SECONDS  stall threshold, 0 disables (default 10)
//
// Signal: SIGUSR1 dumps immediately, whenever the process is still scheduled.

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	watchdogPollEvery   = time.Second
	watchdogRedumpEvery = 5 * time.Minute
	watchdogDefaultStop = 10 * time.Second
)

var (
	watchdogOnce     sync.Once
	watchdogLastBeat atomic.Int64 // unix nanos, written by _process
	watchdogDumpMu   sync.Mutex   // serialises dump writes
	watchdogStarted  = time.Now()
	watchdogPath     string
	watchdogStall    = watchdogDefaultStop
)

// watchdogBeat records that the game loop is alive. Called from _process.
func watchdogBeat() {
	watchdogLastBeat.Store(time.Now().UnixNano())
}

// startWatchdog launches the watchdog once per process.
func startWatchdog() {
	watchdogOnce.Do(func() {
		watchdogPath = os.Getenv("GODOT_GO_WATCHDOG_FILE")
		if watchdogPath == "" {
			watchdogPath = filepath.Join(os.TempDir(), fmt.Sprintf("godot-go-watchdog-%d.log", os.Getpid()))
		}
		if v := os.Getenv("GODOT_GO_WATCHDOG_STALL_SECONDS"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n >= 0 {
				watchdogStall = time.Duration(n) * time.Second
			}
		}

		watchdogLastBeat.Store(time.Now().UnixNano())
		header := fmt.Sprintf("godot-go watchdog: pid %d, started %s, stall threshold %s\n",
			os.Getpid(), watchdogStarted.Format(time.RFC3339), watchdogStall)
		if err := os.WriteFile(watchdogPath, []byte(header), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "godot-go watchdog: cannot write %s: %v\n", watchdogPath, err)
			return
		}
		fmt.Fprintf(os.Stderr, "godot-go watchdog: %s (SIGUSR1 for an immediate dump)\n", watchdogPath)

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGUSR1)

		go func() {
			// A dedicated OS thread gives the watchdog a chance to run while
			// the engine's main thread is parked inside a callback.
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()

			var lastDump time.Time
			ticker := time.NewTicker(watchdogPollEvery)
			defer ticker.Stop()
			for {
				select {
				case <-sigCh:
					lastDump = time.Now()
					watchdogDump("SIGUSR1")

				case <-ticker.C:
					if watchdogStall <= 0 {
						continue
					}
					age := time.Since(time.Unix(0, watchdogLastBeat.Load()))
					if age < watchdogStall {
						continue
					}
					if !lastDump.IsZero() && time.Since(lastDump) < watchdogRedumpEvery {
						continue
					}
					lastDump = time.Now()
					watchdogDump(fmt.Sprintf("no _process heartbeat for %s", age.Round(time.Millisecond)))
				}
			}
		}()
	})
}

// watchdogDump appends one snapshot: process/loop state, kernel thread state,
// then every goroutine stack.
func watchdogDump(reason string) {
	watchdogDumpMu.Lock()
	defer watchdogDumpMu.Unlock()

	now := time.Now()
	beat := time.Unix(0, watchdogLastBeat.Load())

	var b strings.Builder
	fmt.Fprintf(&b, "\n===== godot-go watchdog: %s =====\n", reason)
	fmt.Fprintf(&b, "time:        %s\n", now.Format(time.RFC3339Nano))
	fmt.Fprintf(&b, "pid:         %d\n", os.Getpid())
	fmt.Fprintf(&b, "uptime:      %s\n", now.Sub(watchdogStarted).Round(time.Millisecond))
	fmt.Fprintf(&b, "last beat:   %s ago\n", now.Sub(beat).Round(time.Millisecond))
	fmt.Fprintf(&b, "goroutines:  %d\n", runtime.NumGoroutine())

	b.WriteString("\n--- kernel threads (/proc/self/task) ---\n")
	writeThreadStates(&b)
	b.WriteString("\n--- goroutine stacks ---\n")
	b.Write(watchdogStacks())
	b.WriteString("\n===== end watchdog dump =====\n")

	f, err := os.OpenFile(watchdogPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.WriteString(b.String())
}

// writeThreadStates lists every kernel thread with its state and wait channel.
// This is the half Go cannot see: on a hang it shows whether the engine's main
// thread is parked in a futex, a driver ioctl, or a Wayland read.
func writeThreadStates(b *strings.Builder) {
	entries, err := os.ReadDir("/proc/self/task")
	if err != nil {
		fmt.Fprintf(b, "unavailable: %v\n", err)
		return
	}
	type tstate struct {
		tid          int
		name, state  string
		wchan        string
		utime, stime uint64
	}
	rows := make([]tstate, 0, len(entries))
	for _, e := range entries {
		tid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		dir := filepath.Join("/proc/self/task", e.Name())
		row := tstate{tid: tid}

		if raw, err := os.ReadFile(filepath.Join(dir, "comm")); err == nil {
			row.name = strings.TrimSpace(string(raw))
		}
		if raw, err := os.ReadFile(filepath.Join(dir, "wchan")); err == nil {
			row.wchan = strings.TrimSpace(string(raw))
		}
		if raw, err := os.ReadFile(filepath.Join(dir, "stat")); err == nil {
			// comm may contain spaces and is parenthesised; fields after the
			// closing paren are positional: state is 1st, utime/stime 14th/15th.
			if i := strings.LastIndexByte(string(raw), ')'); i >= 0 {
				fields := strings.Fields(string(raw)[i+1:])
				if len(fields) > 15 {
					row.state = fields[0]
					row.utime, _ = strconv.ParseUint(fields[13], 10, 64)
					row.stime, _ = strconv.ParseUint(fields[14], 10, 64)
				}
			}
		}
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].tid < rows[j].tid })
	for _, r := range rows {
		fmt.Fprintf(b, "%7d  %-20s state=%s  wchan=%-32s cpu=%d.%02ds\n",
			r.tid, r.name, r.state, r.wchan, (r.utime+r.stime)/100, (r.utime+r.stime)%100)
	}
}

// watchdogStacks returns runtime.Stack(all=true), growing the buffer until the
// runtime reports it did not truncate.
func watchdogStacks() []byte {
	size := 1 << 20
	for {
		buf := make([]byte, size)
		n := runtime.Stack(buf, true)
		if n < len(buf) {
			return buf[:n]
		}
		size *= 2
	}
}

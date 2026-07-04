package cli

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"vexil/internal/app"
	"vexil/internal/config"
	"vexil/internal/discovery"
	"vexil/internal/history"
	"vexil/internal/i18n"
	"vexil/internal/protocol"
	"vexil/internal/util"
)

func Run(args []string) {
	if len(args) < 2 {
		cfg := config.LoadAppConfig()
		fmt.Print(i18n.T(cfg.Language, "usage"))
		return
	}

	cfg := config.LoadAppConfig()
	cfg.Apply()
	application := app.New(cfg)
	lang := cfg.Language
	subArgs := args[2:]

	switch args[1] {
	case "send":
		sendCmd(application, subArgs, lang)
	case "recv":
		recvCmd(application, subArgs, cfg, lang)
	case "discover":
		discoverCmd(application, subArgs, lang)
	case "history":
		historyCmd(subArgs, lang)
	case "name":
		nameCmd(subArgs, cfg, lang)
	case "lang":
		langCmd(subArgs, cfg, lang)
	default:
		fmt.Print(i18n.T(lang, "usage"))
	}
}

func sendCmd(app *app.Application, args []string, lang string) {
	if len(args) < 2 {
		fmt.Println(i18n.T(lang, "send_usage"))
		return
	}
	parts := strings.Split(args[0], ":")
	if len(parts) != 2 {
		fmt.Println("invalid host:port") // 简单错误保留原文
		return
	}
	port, err := strconv.Atoi(parts[1])
	if err != nil || port < 1 || port > 65535 {
		fmt.Printf(i18n.T(lang, "invalid_port")+"\n", parts[1])
		return
	}

	paths := args[1:]
	if len(paths) == 0 {
		fmt.Println(i18n.T(lang, "no_files"))
		return
	}

	host := parts[0]
	if net.ParseIP(host) == nil && host != "localhost" {
		fmt.Printf(i18n.T(lang, "searching_device"), host)
		devices, _ := app.DiscoverDevices(3 * time.Second)
		for _, dev := range devices {
			if strings.EqualFold(dev.Name, host) || dev.IP == host {
				host = dev.IP
				port = dev.Port
				fmt.Printf(i18n.T(lang, "device_found"), dev.Name, dev.IP, dev.Port)
				break
			}
		}
	}

	if host == parts[0] && net.ParseIP(host) == nil && host != "localhost" {
		fmt.Println(i18n.T(lang, "unresolvable_host"))
		return
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	hostname, _ := os.Hostname()
	taskID, err := app.StartSend(host, port, paths, hostname)
	if err != nil {
		fmt.Printf(i18n.T(lang, "task_error")+"\n", err)
		os.Exit(1)
	}

	task := app.Task(taskID)
	if task == nil {
		fmt.Println(i18n.T(lang, "task_create_failed"))
		os.Exit(1)
	}

	done := make(chan struct{})
	go func() {
		for ev := range task.Events() {
			switch ev.Type {
			case protocol.EventProgress:
				barLen := int(ev.Percent / 100 * 30)
				if barLen < 0 {
					barLen = 0
				}
				if barLen > 30 {
					barLen = 30
				}
				bar := strings.Repeat("█", barLen)
				empty := strings.Repeat("░", 30-barLen)
				fmt.Printf("\r  %s/%s [%s%s] %.1f%% %s %s          ",
					util.FormatSize(ev.Sent), util.FormatSize(ev.Total), bar, empty,
					ev.Percent, util.FormatSpeed(ev.SpeedMBps*1024*1024),
					util.FormatETA(ev.ETA))
			case protocol.EventError:
				fmt.Printf("\nerror: %v\n", ev.Error)
				close(done)
				return
			case protocol.EventState:
				if ev.State == protocol.TaskCompleted {
					fmt.Print(i18n.T(lang, "done"))
					close(done)
					return
				}
				if ev.State == protocol.TaskCancelled {
					fmt.Print(i18n.T(lang, "cancelled"))
					close(done)
					return
				}
				if ev.State == protocol.TaskFailed {
					fmt.Print(i18n.T(lang, "failed"))
					close(done)
					return
				}
			}
		}
		close(done)
	}()

	select {
	case <-done:
	case sig := <-sigCh:
		fmt.Printf(i18n.T(lang, "signal_received"), sig)
		app.CancelTask(taskID)
		for range task.Events() {
		}
	}
}

func recvCmd(app *app.Application, args []string, cfg *config.AppConfig, lang string) {
	if len(args) < 1 {
		fmt.Println(i18n.T(lang, "recv_usage"))
		return
	}
	port, err := strconv.Atoi(args[0])
	if err != nil || port < 1 || port > 65535 {
		fmt.Printf(i18n.T(lang, "invalid_port")+"\n", args[0])
		return
	}
	saveDir := "."
	if len(args) > 1 {
		saveDir = args[1]
	}
	os.MkdirAll(saveDir, 0755)

	deviceName := cfg.DeviceName
	udpDisc := discovery.NewUDPDiscoveryWithName(deviceName)
	if err := udpDisc.Start(port); err != nil {
		fmt.Printf(i18n.T(lang, "udp_start_fail"), err)
	}
	defer udpDisc.Stop()

	mdnsDisc := discovery.NewMDNSDiscoveryWithName(deviceName)
	if err := mdnsDisc.Start(port); err != nil {
		fmt.Printf(i18n.T(lang, "mdns_start_fail"), err)
	}
	defer mdnsDisc.Stop()

	fmt.Printf(i18n.T(lang, "listening"), port, saveDir)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	taskID, err := app.StartRecv(port, saveDir)
	if err != nil {
		fmt.Printf(i18n.T(lang, "task_error")+"\n", err)
		os.Exit(1)
	}

	task := app.Task(taskID)
	if task == nil {
		fmt.Println(i18n.T(lang, "task_create_failed"))
		os.Exit(1)
	}

	done := make(chan struct{})
	go func() {
		for ev := range task.Events() {
			switch ev.Type {
			case protocol.EventProgress:
				barLen := int(ev.Percent / 100 * 30)
				if barLen < 0 {
					barLen = 0
				}
				if barLen > 30 {
					barLen = 30
				}
				bar := strings.Repeat("█", barLen)
				empty := strings.Repeat("░", 30-barLen)
				fmt.Printf("\r  %s/%s [%s%s] %.1f%% %s %s          ",
					util.FormatSize(ev.Sent), util.FormatSize(ev.Total), bar, empty,
					ev.Percent, util.FormatSpeed(ev.SpeedMBps*1024*1024),
					util.FormatETA(ev.ETA))
			case protocol.EventError:
				fmt.Printf("\nerror: %v\n", ev.Error)
				close(done)
				return
			case protocol.EventState:
				if ev.State == protocol.TaskCompleted {
					fmt.Print(i18n.T(lang, "done"))
					close(done)
					return
				}
				if ev.State == protocol.TaskCancelled {
					fmt.Print(i18n.T(lang, "cancelled"))
					close(done)
					return
				}
				if ev.State == protocol.TaskFailed {
					fmt.Print(i18n.T(lang, "failed"))
					close(done)
					return
				}
			}
		}
		close(done)
	}()

	select {
	case <-done:
	case sig := <-sigCh:
		fmt.Printf(i18n.T(lang, "signal_received"), sig)
		app.CancelTask(taskID)
		for range task.Events() {
		}
	}
}

func discoverCmd(app *app.Application, args []string, lang string) {
	timeout := 3 * time.Second
	if len(args) > 0 {
		if sec, err := strconv.Atoi(args[0]); err == nil && sec > 0 {
			timeout = time.Duration(sec) * time.Second
		}
	}

	fmt.Printf(i18n.T(lang, "discovering"), timeout)
	devices, err := app.DiscoverDevices(timeout)
	if err != nil {
		fmt.Printf(i18n.T(lang, "discover_fail"), err)
		return
	}

	if len(devices) == 0 {
		fmt.Print(i18n.T(lang, "no_devices"))
		return
	}

	fmt.Printf("  %-4s %-20s %-16s %-6s\n", "", "主机名", "IP", "端口")
	fmt.Println("  " + strings.Repeat("─", 50))
	for i, d := range devices {
		fmt.Printf("  [%d] %-20s %-16s %-6d\n", i+1, d.Name, d.IP, d.Port)
	}
	fmt.Println()
}

func historyCmd(args []string, lang string) {
	if len(args) > 0 && args[0] == "clear" {
		if len(args) == 1 {
			if err := history.Clear(); err != nil {
				fmt.Printf(i18n.T(lang, "clear_fail"), err)
			} else {
				fmt.Print(i18n.T(lang, "history_cleared"))
			}
			return
		}
		index, err := strconv.Atoi(args[1])
		if err != nil || index < 1 {
			fmt.Printf(i18n.T(lang, "invalid_index"), args[1])
			return
		}
		if err := history.Delete(index); err != nil {
			fmt.Printf(i18n.T(lang, "delete_fail"), err)
		} else {
			fmt.Printf(i18n.T(lang, "history_deleted"), index)
		}
		return
	}

	limit := 20
	if len(args) > 0 {
		if n, err := strconv.Atoi(args[0]); err == nil && n > 0 {
			limit = n
		}
	}

	entries, err := history.List(limit)
	if err != nil || len(entries) == 0 {
		fmt.Print(i18n.T(lang, "no_history"))
		return
	}

	fmt.Printf(i18n.T(lang, "recent_history"), len(entries))
	for i, e := range entries {
		fmt.Println(history.FormatEntry(e, i, lang))
	}
	fmt.Println()
}

func nameCmd(args []string, cfg *config.AppConfig, lang string) {
	if len(args) < 1 {
		if cfg.DeviceName != "" {
			fmt.Printf(i18n.T(lang, "current_device_name"), cfg.DeviceName)
		} else {
			hostname, _ := os.Hostname()
			fmt.Printf(i18n.T(lang, "device_name_default"), hostname)
		}
		return
	}

	cfg.DeviceName = args[0]
	if err := cfg.Save(); err != nil {
		fmt.Printf(i18n.T(lang, "name_save_fail"), err)
		return
	}
	fmt.Printf(i18n.T(lang, "device_name_updated"), args[0])
}

func langCmd(args []string, cfg *config.AppConfig, lang string) {
	if len(args) < 1 {
		langName := "中文"
		if lang == "en" {
			langName = "English"
		}
		fmt.Printf(i18n.T(lang, "current_lang"), langName)
		return
	}

	newLang := args[0]
	if newLang != "zh" && newLang != "en" {
		fmt.Print(i18n.T(lang, "lang_invalid"))
		return
	}

	cfg.Language = newLang
	if err := cfg.Save(); err != nil {
		fmt.Printf(i18n.T(lang, "name_save_fail"), err)
		return
	}
	fmt.Printf(i18n.T(newLang, "lang_updated"), newLang)
}
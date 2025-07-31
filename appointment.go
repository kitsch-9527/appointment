package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// 全局配置变量
var (
	cfg = struct {
		requestURL       string
		sleepTime        time.Duration
		maxIndex         int
		maxLoopDuration  time.Duration
		requestTimeout   time.Duration
		timeSlot         string
		token            string
		lineName         string
		stationName      string
		snapshotTimeSlot string
		dateOffset       int
	}{
		// 默认配置
		requestURL:       "https://webapi.mybti.cn/Appointment/CreateAppointment",
		sleepTime:        1 * time.Second,
		maxIndex:         15,
		maxLoopDuration:  2 * time.Minute,
		requestTimeout:   10 * time.Second,
		timeSlot:         "0820-0830",
		lineName:         "昌平线",
		stationName:      "沙河站",
		snapshotTimeSlot: "0630-0930",
		dateOffset:       1, // 默认预约明天
	}
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "appointment",
		Short: "地铁预约工具",
		Long:  "用于预约地铁进站时段的命令行工具",
		Run:   runAppointment,
	}

	// 配置参数
	rootCmd.Flags().StringVarP(&cfg.token, "token", "t", "", "预约系统授权token (必填)")
	rootCmd.Flags().BoolP("loop", "p", false, "启用2分钟循环预约模式")
	rootCmd.Flags().DurationVarP(&cfg.sleepTime, "sleep", "s", cfg.sleepTime, "重试间隔时间")
	rootCmd.Flags().IntVarP(&cfg.maxIndex, "max-retry", "r", cfg.maxIndex, "最大重试次数(有限模式)")
	rootCmd.Flags().DurationVarP(&cfg.maxLoopDuration, "loop-duration", "d", cfg.maxLoopDuration, "循环模式最大时长")
	rootCmd.Flags().DurationVarP(&cfg.requestTimeout, "timeout", "o", cfg.requestTimeout, "请求超时时间")

	// 预约参数
	rootCmd.Flags().StringVarP(&cfg.timeSlot, "time-slot", "l", cfg.timeSlot, "预约具体时段(如:0820-0830)")
	rootCmd.Flags().StringVarP(&cfg.lineName, "line", "L", cfg.lineName, "地铁线路名称")
	rootCmd.Flags().StringVarP(&cfg.stationName, "station", "n", cfg.stationName, "车站名称")

	// 标记必填参数
	_ = rootCmd.MarkFlagRequired("token")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runAppointment(cmd *cobra.Command, args []string) {
	// 获取命令行参数
	loopMode, _ := cmd.Flags().GetBool("loop")

	// 初始化预约参数
	currentDate := getDate(cfg.dateOffset)

	row := ReserveRequest{
		LineName:           cfg.lineName,
		SnapshotWeekOffset: 0,
		StationName:        cfg.stationName,
		EnterDate:          currentDate,
		SnapshotTimeSlot:   cfg.snapshotTimeSlot,
		TimeSlot:           cfg.timeSlot,
	}
	// 根据模式执行
	if loopMode {
		fmt.Println("启用循环预约模式...")
		_, err := LoopWithTimeLimit(cfg.token, row)
		if err != nil {
			fmt.Printf("循环预约结束: %v\n", err)
		}
	} else {
		fmt.Println("启用有限重试模式...")
		_, err := RetryWithLimit(cfg.token, row)
		if err != nil {
			fmt.Printf("有限重试结束: %v\n", err)
		}
	}
}

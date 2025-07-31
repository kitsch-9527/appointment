package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// 预约请求体结构
type ReserveRequest struct {
	LineName           string `json:"lineName"`
	SnapshotWeekOffset int    `json:"snapshotWeekOffset"`
	StationName        string `json:"stationName"`
	EnterDate          string `json:"enterDate"`
	SnapshotTimeSlot   string `json:"snapshotTimeSlot"`
	TimeSlot           string `json:"timeSlot"`
}

// 响应结构体
type ReserveResponse struct {
	Balance         int    `json:"balance"`         // 余额信息
	AppointmentId   string `json:"appointmentId"`   // 预约ID
	StationEntrance string `json:"stationEntrance"` // 入口信息（字符串类型，如"A2口"）
	Message         string `json:"message"`         // 响应消息
	StatusCode      int    `json:"-"`
}

// 获取指定天数后的日期（格式：YYYYMMDD）
func getDate(day int) string {
	dd := time.Now().AddDate(0, 0, day)
	return dd.Format("20060102")
}

// 有限重试模式（最多重试 maxIndex 次）
func RetryWithLimit(token string, reqData ReserveRequest) (ReserveResponse, error) {
	var lastResp ReserveResponse
	var lastErr error

	for i := 0; i < cfg.maxIndex; i++ {
		fmt.Printf("\n第 %d 次预约请求...\n", i+1)
		success, resp, err := sendReservationRequest(token, reqData)

		// 保存最后一次响应和错误
		lastResp = resp
		lastErr = err

		if err != nil {
			fmt.Printf("请求失败: %v\n", err)
		} else {
			fmt.Printf("响应结果: %+v\n", resp)
			if success {
				fmt.Printf("%s 预约进站成功\n", reqData.EnterDate)
				return resp, nil
			}
		}

		// 不是最后一次尝试才需要等待
		if i < cfg.maxIndex-1 {
			fmt.Printf("预约失败，将在 %v 后重试...\n", cfg.sleepTime)
			time.Sleep(cfg.sleepTime)
		}
	}

	return lastResp, fmt.Errorf("已达到最大重试次数(%d次)，最后错误: %v", cfg.maxIndex, lastErr)
}

// 时间限制循环模式
func LoopWithTimeLimit(token string, reqData ReserveRequest) (ReserveResponse, error) {
	startTime := time.Now()
	var lastResp ReserveResponse
	var lastErr error

	for {
		// 检查是否超时
		if time.Since(startTime) > cfg.maxLoopDuration {
			return lastResp, fmt.Errorf("超过最大循环时长(%v)，未预约成功", cfg.maxLoopDuration)
		}

		// 执行预约请求
		success, resp, err := sendReservationRequest(token, reqData)
		lastResp = resp
		lastErr = err

		if err != nil {
			fmt.Printf("请求错误: %v\n", err)
		} else {
			fmt.Printf("响应结果: %+v\n", resp)
			if success {
				fmt.Printf("%s 预约进站成功\n", reqData.EnterDate)
				return resp, nil
			}
		}

		fmt.Printf("预约失败%s，%v 后再次尝试...\n", lastErr, cfg.sleepTime)
		time.Sleep(cfg.sleepTime)
	}
}

// 发送预约请求
func sendReservationRequest(token string, reqData ReserveRequest) (bool, ReserveResponse, error) {
	// 序列化请求体
	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return false, ReserveResponse{}, fmt.Errorf("请求体序列化失败: %w", err)
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest(http.MethodPost, cfg.requestURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return false, ReserveResponse{}, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	setRequestHeaders(req, token)
	// 打印请求体
	fmt.Printf("请求体: %s\n", string(jsonData))
	// 发送请求
	client := &http.Client{Timeout: cfg.requestTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return false, ReserveResponse{}, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, ReserveResponse{}, fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析响应
	var reserveResp ReserveResponse
	reserveResp.StatusCode = resp.StatusCode // 保存HTTP状态码

	if err := json.Unmarshal(body, &reserveResp); err != nil {
		// 解析失败时，将原始响应作为消息保存
		reserveResp.Message = fmt.Sprintf("解析响应失败: %v, 原始响应: %s", err, string(body))
		return false, reserveResp, nil
	}

	// 检查HTTP状态码是否成功
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return false, reserveResp, fmt.Errorf("请求返回非成功状态码: %d", resp.StatusCode)
	}

	// 判断是否预约成功
	success := reserveResp.AppointmentId != "" && reserveResp.StationEntrance != ""
	return success, reserveResp, nil
}

// 设置请求头
func setRequestHeaders(req *http.Request, token string) {
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("authorization", token)
	req.Header.Set("content-type", "application/json;charset=UTF-8")
	req.Header.Set("sec-ch-ua", `"Not_A Brand";v="99", "Google Chrome";v="109", "Chromium";v="109"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", "macOS")
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-site")
	req.Header.Set("Referer", "https://webui.mybti.cn/")
	req.Header.Set("Referrer-Policy", "strict-origin-when-cross-origin")
}

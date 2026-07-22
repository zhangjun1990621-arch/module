package handler

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"iot-platform/internal/database"
	"iot-platform/internal/model"
)

// OTAHandler OTA 升级处理器
type OTAHandler struct{}

func NewOTAHandler() *OTAHandler {
	return &OTAHandler{}
}

// ================ 固件管理 ================

// ListFirmwares 固件列表
func (h *OTAHandler) ListFirmwares(c *gin.Context) {
	pdb, ok := getPlatformTx(c)
	if !ok {
		return
	}
	committed := false
	defer func() {
		if !committed {
			pdb.Rollback()
		}
	}()

	var firmwares []model.Firmware
	if err := pdb.Order("upload_time DESC").Find(&firmwares).Error; err != nil {
		fail(c, http.StatusInternalServerError, "查询固件列表失败: "+err.Error())
		return
	}
	if firmwares == nil {
		firmwares = make([]model.Firmware, 0)
	}

	if err := pdb.Commit().Error; err != nil {
		fail(c, http.StatusInternalServerError, "提交事务失败: "+err.Error())
		return
	}
	committed = true
	success(c, firmwares)
}

// UploadFirmware 上传固件
func (h *OTAHandler) UploadFirmware(c *gin.Context) {
	pdb, ok := getPlatformTx(c)
	if !ok {
		return
	}
	committed := false
	defer func() {
		if !committed {
			pdb.Rollback()
		}
	}()

	file, err := c.FormFile("file")
	if err != nil {
		fail(c, http.StatusBadRequest, "请选择固件文件")
		return
	}

	name := c.PostForm("name")
	if name == "" {
		name = file.Filename
	}

	firmware := model.Firmware{
		Name:       name,
		Version:    c.PostForm("version"),
		FilePath:   fmt.Sprintf("/uploads/%s", file.Filename),
		FileSize:   file.Size,
		DeviceType: c.PostForm("deviceType"),
		UploadTime: time.Now(),
	}

	if err := pdb.Create(&firmware).Error; err != nil {
		fail(c, http.StatusInternalServerError, "保存固件记录失败: "+err.Error())
		return
	}

	if err := pdb.Commit().Error; err != nil {
		fail(c, http.StatusInternalServerError, "提交事务失败: "+err.Error())
		return
	}
	committed = true
	c.JSON(http.StatusCreated, APIResponse{Code: 0, Message: "success", Data: firmware})
}

// DeleteFirmware 删除固件
func (h *OTAHandler) DeleteFirmware(c *gin.Context) {
	pdb, ok := getPlatformTx(c)
	if !ok {
		return
	}
	committed := false
	defer func() {
		if !committed {
			pdb.Rollback()
		}
	}()

	id := c.Param("id")

	// 检查是否被升级任务引用
	var taskCount int64
	pdb.Model(&model.OTATask{}).Where("firmware_id = ?", id).Count(&taskCount)
	if taskCount > 0 {
		fail(c, http.StatusBadRequest, fmt.Sprintf("该固件被 %d 个升级任务引用，无法删除", taskCount))
		return
	}

	result := pdb.Where("id = ?", id).Delete(&model.Firmware{})
	if result.Error != nil {
		fail(c, http.StatusInternalServerError, "删除固件失败: "+result.Error.Error())
		return
	}
	if result.RowsAffected == 0 {
		fail(c, http.StatusNotFound, "固件不存在")
		return
	}

	if err := pdb.Commit().Error; err != nil {
		fail(c, http.StatusInternalServerError, "提交事务失败: "+err.Error())
		return
	}
	committed = true
	success(c, gin.H{"message": "固件已删除"})
}

// ================ 升级任务 ================

// ListTasks 任务列表
func (h *OTAHandler) ListTasks(c *gin.Context) {
	pdb, ok := getPlatformTx(c)
	if !ok {
		return
	}
	committed := false
	defer func() {
		if !committed {
			pdb.Rollback()
		}
	}()

	var tasks []model.OTATask
	if err := pdb.Preload("Firmware").Order("created_at DESC").Find(&tasks).Error; err != nil {
		fail(c, http.StatusInternalServerError, "查询任务列表失败: "+err.Error())
		return
	}
	if tasks == nil {
		tasks = make([]model.OTATask, 0)
	}

	if err := pdb.Commit().Error; err != nil {
		fail(c, http.StatusInternalServerError, "提交事务失败: "+err.Error())
		return
	}
	committed = true
	success(c, tasks)
}

// GetTask 任务详情
func (h *OTAHandler) GetTask(c *gin.Context) {
	pdb, ok := getPlatformTx(c)
	if !ok {
		return
	}
	committed := false
	defer func() {
		if !committed {
			pdb.Rollback()
		}
	}()

	id := c.Param("id")
	var task model.OTATask
	if err := pdb.Preload("Firmware").Preload("TaskDevices").Where("id = ?", id).First(&task).Error; err != nil {
		fail(c, http.StatusNotFound, "任务不存在")
		return
	}

	if err := pdb.Commit().Error; err != nil {
		fail(c, http.StatusInternalServerError, "提交事务失败: "+err.Error())
		return
	}
	committed = true
	success(c, task)
}

// CreateTaskRequest 创建任务请求体
type CreateTaskRequest struct {
	FirmwareID uint     `json:"firmwareId" binding:"required"`
	DeviceIDs  []string `json:"deviceIds" binding:"required"`
}

// CreateTask 创建升级任务
func (h *OTAHandler) CreateTask(c *gin.Context) {
	pdb, ok := getPlatformTx(c)
	if !ok {
		return
	}
	committed := false
	defer func() {
		if !committed {
			pdb.Rollback()
		}
	}()

	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	if len(req.DeviceIDs) == 0 {
		fail(c, http.StatusBadRequest, "请选择至少一台设备")
		return
	}

	// 验证固件存在
	var firmware model.Firmware
	if err := pdb.Where("id = ?", req.FirmwareID).First(&firmware).Error; err != nil {
		fail(c, http.StatusBadRequest, "固件不存在")
		return
	}

	// 创建任务
	task := model.OTATask{
		FirmwareID:   req.FirmwareID,
		Firmware:     &firmware,
		Status:       model.OTATaskPending,
		TotalDevices: len(req.DeviceIDs),
		CreatedBy:    c.GetString("username"),
	}

	if err := pdb.Create(&task).Error; err != nil {
		fail(c, http.StatusInternalServerError, "创建任务失败: "+err.Error())
		return
	}

	// 创建设备明细
	for _, deviceID := range req.DeviceIDs {
		td := model.OTATaskDevice{
			TaskID:   task.ID,
			DeviceID: deviceID,
			Status:   model.OTADevicePending,
		}
		if err := pdb.Create(&td).Error; err != nil {
			fail(c, http.StatusInternalServerError, "创建设备明细失败: "+err.Error())
			return
		}
	}

	if err := pdb.Commit().Error; err != nil {
		fail(c, http.StatusInternalServerError, "提交事务失败: "+err.Error())
		return
	}
	committed = true

	// 启动后台进度推进 goroutine
	platform := c.Param("platform")
	go h.runUpgradeProgress(platform, task.ID)

	c.JSON(http.StatusCreated, APIResponse{Code: 0, Message: "success", Data: task})
}

// PauseTask 暂停任务
func (h *OTAHandler) PauseTask(c *gin.Context) {
	h.updateTaskStatus(c, model.OTATaskPaused, "")
}

// ResumeTask 继续任务
func (h *OTAHandler) ResumeTask(c *gin.Context) {
	pdb, ok := getPlatformTx(c)
	if !ok {
		return
	}
	committed := false
	defer func() {
		if !committed {
			pdb.Rollback()
		}
	}()

	id := c.Param("id")
	var task model.OTATask
	if err := pdb.Where("id = ?", id).First(&task).Error; err != nil {
		fail(c, http.StatusNotFound, "任务不存在")
		return
	}

	if task.Status != model.OTATaskPaused {
		fail(c, http.StatusBadRequest, "只有已暂停的任务才能继续")
		return
	}

	pdb.Model(&task).Update("status", model.OTATaskRunning)
	task.Status = model.OTATaskRunning

	if err := pdb.Commit().Error; err != nil {
		fail(c, http.StatusInternalServerError, "提交事务失败: "+err.Error())
		return
	}
	committed = true
	success(c, task)
}

// CancelTask 取消任务
func (h *OTAHandler) CancelTask(c *gin.Context) {
	pdb, ok := getPlatformTx(c)
	if !ok {
		return
	}
	committed := false
	defer func() {
		if !committed {
			pdb.Rollback()
		}
	}()

	id := c.Param("id")
	var task model.OTATask
	if err := pdb.Where("id = ?", id).First(&task).Error; err != nil {
		fail(c, http.StatusNotFound, "任务不存在")
		return
	}

	if task.Status == model.OTATaskCompleted || task.Status == model.OTATaskCancelled {
		fail(c, http.StatusBadRequest, "任务已结束，无法取消")
		return
	}

	now := time.Now()
	pdb.Model(&task).Updates(map[string]interface{}{
		"status":       model.OTATaskCancelled,
		"completed_at": now,
		"end_reason":   model.OTAEndCancelled,
	})

	// 将所有未完成的设备标记为失败
	pdb.Model(&model.OTATaskDevice{}).
		Where("task_id = ? AND status IN ?", task.ID, []string{model.OTADevicePending, model.OTADeviceUpgrading}).
		Updates(map[string]interface{}{
			"status":     model.OTADeviceFailed,
			"error_msg":  "任务已取消",
			"updated_at": now,
		})

	h.recalcProgress(pdb, task.ID)
	task.Status = model.OTATaskCancelled

	if err := pdb.Commit().Error; err != nil {
		fail(c, http.StatusInternalServerError, "提交事务失败: "+err.Error())
		return
	}
	committed = true
	success(c, task)
}

// CompleteTask 手动结束任务（进度停在当前位置）
func (h *OTAHandler) CompleteTask(c *gin.Context) {
	pdb, ok := getPlatformTx(c)
	if !ok {
		return
	}
	committed := false
	defer func() {
		if !committed {
			pdb.Rollback()
		}
	}()

	id := c.Param("id")
	var task model.OTATask
	if err := pdb.Where("id = ?", id).First(&task).Error; err != nil {
		fail(c, http.StatusNotFound, "任务不存在")
		return
	}

	if task.Status != model.OTATaskRunning && task.Status != model.OTATaskPaused {
		fail(c, http.StatusBadRequest, "只有运行中或已暂停的任务才能手动结束")
		return
	}

	now := time.Now()
	// 将未完成的设备标记为失败
	pdb.Model(&model.OTATaskDevice{}).
		Where("task_id = ? AND status IN ?", task.ID, []string{model.OTADevicePending, model.OTADeviceUpgrading}).
		Updates(map[string]interface{}{
			"status":     model.OTADeviceFailed,
			"error_msg":  "手动结束任务",
			"updated_at": now,
		})

	h.recalcProgress(pdb, task.ID)

	// 确定结束原因
	var task2 model.OTATask
	pdb.Where("id = ?", task.ID).First(&task2)

	endReason := model.OTAEndManualStop
	if task2.SuccessCount == task2.TotalDevices {
		endReason = model.OTAEndAllSuccess
	} else if task2.SuccessCount > 0 {
		endReason = model.OTAEndPartialFail
	} else {
		endReason = model.OTAEndAllFailed
	}

	pdb.Model(&task).Updates(map[string]interface{}{
		"status":       model.OTATaskCompleted,
		"completed_at": now,
		"end_reason":   endReason,
	})

	task2.Status = model.OTATaskCompleted
	task2.EndReason = endReason

	if err := pdb.Commit().Error; err != nil {
		fail(c, http.StatusInternalServerError, "提交事务失败: "+err.Error())
		return
	}
	committed = true
	success(c, task2)
}

// DeleteTask 删除任务（软删除）
func (h *OTAHandler) DeleteTask(c *gin.Context) {
	pdb, ok := getPlatformTx(c)
	if !ok {
		return
	}
	committed := false
	defer func() {
		if !committed {
			pdb.Rollback()
		}
	}()

	id := c.Param("id")
	var task model.OTATask
	if err := pdb.Where("id = ?", id).First(&task).Error; err != nil {
		fail(c, http.StatusNotFound, "任务不存在")
		return
	}

	// 运行中/暂停的任务不能直接删除，必须先取消
	if task.Status == model.OTATaskRunning || task.Status == model.OTATaskPaused {
		fail(c, http.StatusBadRequest, "任务正在运行或已暂停，请先取消再删除")
		return
	}

	// 删除设备明细
	pdb.Where("task_id = ?", task.ID).Delete(&model.OTATaskDevice{})
	// 删除任务
	pdb.Where("id = ?", id).Delete(&model.OTATask{})

	if err := pdb.Commit().Error; err != nil {
		fail(c, http.StatusInternalServerError, "提交事务失败: "+err.Error())
		return
	}
	committed = true
	success(c, gin.H{"message": "任务已删除"})
}

// RetryFailedDevices 重试失败设备
func (h *OTAHandler) RetryFailedDevices(c *gin.Context) {
	pdb, ok := getPlatformTx(c)
	if !ok {
		return
	}
	committed := false
	defer func() {
		if !committed {
			pdb.Rollback()
		}
	}()

	id := c.Param("id")
	var task model.OTATask
	if err := pdb.Where("id = ?", id).First(&task).Error; err != nil {
		fail(c, http.StatusNotFound, "任务不存在")
		return
	}

	if task.Status != model.OTATaskCompleted && task.Status != model.OTATaskCancelled {
		fail(c, http.StatusBadRequest, "只有已完成的任务才能重试失败设备")
		return
	}

	// 重置失败设备状态为 pending
	pdb.Model(&model.OTATaskDevice{}).
		Where("task_id = ? AND status = ?", task.ID, model.OTADeviceFailed).
		Updates(map[string]interface{}{
			"status":     model.OTADevicePending,
			"error_msg":  "",
			"updated_at": time.Now(),
		})

	// 重置任务状态
	pdb.Model(&task).Updates(map[string]interface{}{
		"status":       model.OTATaskRunning,
		"completed_at": nil,
		"end_reason":   "",
	})

	if err := pdb.Commit().Error; err != nil {
		fail(c, http.StatusInternalServerError, "提交事务失败: "+err.Error())
		return
	}
	committed = true

	// 重新启动进度推进
	platform := c.Param("platform")
	go h.runUpgradeProgress(platform, task.ID)

	success(c, gin.H{"message": "重试已启动"})
}

// ================ 内部方法 ================

// updateTaskStatus 通用状态更新（用于暂停）
func (h *OTAHandler) updateTaskStatus(c *gin.Context, status, endReason string) {
	pdb, ok := getPlatformTx(c)
	if !ok {
		return
	}
	committed := false
	defer func() {
		if !committed {
			pdb.Rollback()
		}
	}()

	id := c.Param("id")
	var task model.OTATask
	if err := pdb.Where("id = ?", id).First(&task).Error; err != nil {
		fail(c, http.StatusNotFound, "任务不存在")
		return
	}

	updates := map[string]interface{}{"status": status}
	if endReason != "" {
		updates["end_reason"] = endReason
		now := time.Now()
		updates["completed_at"] = now
	}

	pdb.Model(&task).Updates(updates)
	task.Status = status

	if err := pdb.Commit().Error; err != nil {
		fail(c, http.StatusInternalServerError, "提交事务失败: "+err.Error())
		return
	}
	committed = true
	success(c, task)
}

// recalcProgress 重新计算任务进度和统计
func (h *OTAHandler) recalcProgress(pdb *gorm.DB, taskID uint) {
	var successCount, failCount int64
	pdb.Model(&model.OTATaskDevice{}).Where("task_id = ? AND status = ?", taskID, model.OTADeviceSuccess).Count(&successCount)
	pdb.Model(&model.OTATaskDevice{}).Where("task_id = ? AND status = ?", taskID, model.OTADeviceFailed).Count(&failCount)

	var task model.OTATask
	pdb.Where("id = ?", taskID).First(&task)

	progress := 0
	if task.TotalDevices > 0 {
		progress = int((successCount + failCount) * 100 / int64(task.TotalDevices))
	}

	pdb.Model(&task).Updates(map[string]interface{}{
		"success_count": successCount,
		"fail_count":    failCount,
		"progress":      progress,
	})
}

// runUpgradeProgress 后台 goroutine：逐设备模拟升级，实时更新进度
func (h *OTAHandler) runUpgradeProgress(platform string, taskID uint) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for {
		// 每次操作获取新的 DB 连接（goroutine 不能复用请求事务）
		pdb, err := database.GetPlatformDB(platform)
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}

		var task model.OTATask
		if err := pdb.Where("id = ?", taskID).First(&task).Error; err != nil {
			pdb.Rollback()
			time.Sleep(2 * time.Second)
			continue
		}

		// 任务已结束，停止 goroutine
		if task.Status == model.OTATaskCompleted || task.Status == model.OTATaskCancelled {
			pdb.Rollback()
			return
		}

		// 暂停中：等待恢复
		if task.Status == model.OTATaskPaused {
			pdb.Rollback()
			time.Sleep(2 * time.Second)
			continue
		}

		// 状态为 pending 时先改为 running
		if task.Status == model.OTATaskPending {
			pdb.Model(&task).Update("status", model.OTATaskRunning)
		}

		// 找到下一个待升级的设备
		var device model.OTATaskDevice
		if err := pdb.Where("task_id = ? AND status = ?", taskID, model.OTADevicePending).
			First(&device).Error; err != nil {

			// 没有更多待处理的设备，标记完成
			h.recalcProgress(pdb, taskID)

			var finalTask model.OTATask
			pdb.Where("id = ?", taskID).First(&finalTask)

			endReason := model.OTAEndAllSuccess
			if finalTask.SuccessCount == 0 {
				endReason = model.OTAEndAllFailed
			} else if finalTask.FailCount > 0 {
				endReason = model.OTAEndPartialFail
			}

			now := time.Now()
			pdb.Model(&finalTask).Updates(map[string]interface{}{
				"status":       model.OTATaskCompleted,
				"completed_at": now,
				"end_reason":   endReason,
			})

			// 收集失败设备列表
			var failedDevs []model.OTATaskDevice
			pdb.Where("task_id = ? AND status = ?", taskID, model.OTADeviceFailed).Find(&failedDevs)
			if len(failedDevs) > 0 {
				failedIDs := make([]string, len(failedDevs))
				for i, fd := range failedDevs {
					failedIDs[i] = fd.DeviceID
				}
				failedJSON, _ := json.Marshal(failedIDs)
				pdb.Model(&finalTask).Update("failed_devices", datatypes.JSON(failedJSON))
			}

			pdb.Commit()
			return
		}

		// 标记设备为升级中
		pdb.Model(&device).Updates(map[string]interface{}{
			"status":     model.OTADeviceUpgrading,
			"updated_at": time.Now(),
		})
		pdb.Commit()

		// 模拟升级耗时（1-3 秒）
		upgradeTime := time.Duration(1000+rng.Intn(2000)) * time.Millisecond
		time.Sleep(upgradeTime)

		// 重新获取连接更新结果
		pdb2, err := database.GetPlatformDB(platform)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		// 模拟升级：100% 成功率（演示环境，确保一次升级成功）
		deviceStatus := model.OTADeviceSuccess
		errMsg := ""

		pdb2.Model(&model.OTATaskDevice{}).Where("id = ?", device.ID).Updates(map[string]interface{}{
			"status":     deviceStatus,
			"error_msg":  errMsg,
			"updated_at": time.Now(),
		})

		// 更新任务进度
		h.recalcProgress(pdb2, taskID)
		pdb2.Commit()
	}
}

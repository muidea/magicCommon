package def

// UploadFileResult 上传文件结果
type UploadFileResult struct {
	Result
	FileToken string `json:"fileToken"`
}

// DownloadFileResult 下载文件结果
type DownloadFileResult struct {
	Result
	RedirectURL string `json:"redirectUrl"`
}

// DeleteFileResult 删除文件结果
type DeleteFileResult Result

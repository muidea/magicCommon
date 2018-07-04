package model

// FileSummary 文件信息
type FileSummary struct {
	FileToken   string `json:"fileToken"`
	FileName    string `json:"fileName"`
	FilePath    string `json:"filePath"`
	UploadDate  string `json:"uploadDate"`
	ReserveFlag int    `json:"reserveFlag"`
}

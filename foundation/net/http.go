package net

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicCommon/foundation/util"
)

// maxBytesReader 限制从底层读取器读取的字节数量不超过指定的最大值
type maxBytesReader struct {
	res               http.ResponseWriter // HTTP响应写入器
	reader            io.ReadCloser       // 底层读取器
	maxBytesRemaining int64               // 剩余最大字节数
	err               error               // 持久性错误
}

func (l *maxBytesReader) Read(buffer []byte) (n int, err error) {
	if l.err != nil {
		return 0, l.err
	}
	if len(buffer) == 0 {
		return 0, nil
	}
	// 如果请求读取的字节数比剩余允许的最大字节数大，
	// 则限制读取缓冲区大小为剩余字节数+1，以检测是否超出限制。
	if int64(len(buffer)) > l.maxBytesRemaining+1 {
		buffer = buffer[:l.maxBytesRemaining+1]
	}
	n, err = l.reader.Read(buffer)

	if int64(n) <= l.maxBytesRemaining {
		l.maxBytesRemaining -= int64(n)
		l.err = err
		return n, err
	}

	n = int(l.maxBytesRemaining)
	l.maxBytesRemaining = 0

	// 服务端和客户端代码都使用 maxBytesReader。
	// 这个 "requestTooLarge" 检查只在服务端代码中使用。
	// 为了防止仅使用HTTP客户端代码的二进制文件（如 cmd/go）也链接HTTP服务器，
	// 不要对服务器 "*response" 类型使用静态类型断言，
	// 而是检查这个接口：
	type requestTooLarger interface {
		requestTooLarge()
	}
	if res, ok := l.res.(requestTooLarger); ok {
		res.requestTooLarge()
	}
	l.err = errors.New("http: request body too large")
	return n, l.err
}

func (l *maxBytesReader) Close() error {
	return l.reader.Close()
}

// GetHTTPRemoteAddress get http remote address
func GetHTTPRemoteAddress(req *http.Request) (ret string) {
	ret = req.Header.Get("x-forwarded-for")
	if ret == "" {
		ret = req.RemoteAddr
	}

	ret = strings.Split(ret, ", ")[0]

	return
}

// GetHTTPRequestBody get http request body
func GetHTTPRequestBody(req *http.Request) (ret []byte, err error) {
	var reader io.Reader = req.Body
	maxFormSize := int64(1<<63 - 1)
	if _, ok := req.Body.(*maxBytesReader); !ok {
		maxFormSize = int64(10 << 20) // 10 MB is a lot of text.
		reader = io.LimitReader(req.Body, maxFormSize+1)
	}

	payload, payloadErr := io.ReadAll(reader)
	if payloadErr != nil {
		err = payloadErr
		log.Errorf("read request body error: %v", err)
		return
	}

	if int64(len(payload)) > maxFormSize {
		err = errors.New("http: request body too large")
		return
	}

	ret = payload
	return
}

func HTTPBodyToFile(req *http.Request, dstFilePath, fileName string) (err error) {
	var reader io.Reader = req.Body
	maxFormSize := int64(1<<63 - 1)
	if _, ok := req.Body.(*maxBytesReader); !ok {
		maxFormSize = int64(10 << 20) // 10 MB is a lot of text.
		reader = io.LimitReader(req.Body, maxFormSize+1)
	}
	// 验证 dstFilePath 是否为合法的目录路径
	if !isValidDirectory(dstFilePath) {
		err = fmt.Errorf("invalid destination directory: %s", dstFilePath)
		log.Errorf("invalid destination directory, err: %s", err.Error())
		return
	}

	// 验证文件名是否合法
	if !isValidFileName(fileName) {
		err = fmt.Errorf("invalid file name: %s", fileName)
		log.Errorf("invalid file name, err: %s", err.Error())
		return
	}

	// 构建目标文件的完整路径
	dstFullFilePath := filepath.Join(dstFilePath, fileName)
	// 创建目标文件
	dstFileHandle, dstFileErr := os.Create(dstFullFilePath)
	if dstFileErr != nil {
		err = dstFileErr
		log.Errorf("create destination file failed, err: %s", err.Error())
		return
	}
	defer dstFileHandle.Close()

	_, err = io.Copy(dstFileHandle, reader)
	if err != nil {
		log.Errorf("copy destination file failed, err: %s", err.Error())
		return
	}
	return
}

// ParseJSONBody 解析http body请求提交的json数据
func ParseJSONBody(req *http.Request, validator util.Validator, param interface{}) error {
	util.ValidatePtr(param)

	if req.Body == nil {
		return errors.New("missing form body")
	}

	contentType, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil {
		log.Errorf("parse content-type error: %v", err)
		return err
	}

	switch {
	case contentType == "application/json":
		payload, payloadErr := GetHTTPRequestBody(req)
		if payloadErr != nil {
			log.Errorf("get http request body error: %v", payloadErr)
			return payloadErr
		}

		err = json.Unmarshal(payload, param)
		if err != nil {
			log.Errorf("unmarshal http request body error: %v", err)
			return err
		}

		if validator != nil {
			err = validator.Validate(param)
			if err != nil {
				log.Errorf("validate http request body error: %v", err)
				return err
			}
		}

	default:
		return errors.New("invalid contentType, contentType:" + contentType)
	}

	return nil
}

var contentType = textproto.CanonicalMIMEHeaderKey("content-type")

func verifyContentType(res http.ResponseWriter) {
	contentVal := res.Header().Get(contentType)
	if contentVal != "" {
		return
	}
	res.Header().Set(contentType, "application/json; charset=utf-8")
}
func PackageHTTPResponse(res http.ResponseWriter, result interface{}) {
	verifyContentType(res)

	if result == nil {
		res.WriteHeader(http.StatusOK)
		return
	}

	block, err := json.Marshal(result)
	if err == nil {
		_, err := res.Write(block)
		if err != nil {
			log.Errorf("write result error: %v", err)
		}
		return
	}

	log.Errorf("marshal result error: %v", err)
	res.WriteHeader(http.StatusExpectationFailed)
}

func PackageHTTPResponseWithStatusCode(res http.ResponseWriter, statusCode int, result interface{}) {
	verifyContentType(res)

	res.WriteHeader(statusCode)
	if result == nil {
		return
	}

	block, err := json.Marshal(result)
	if err == nil {
		_, err := res.Write(block)
		if err != nil {
			log.Errorf("write result error: %v", err)
		}
		return
	}

	log.Errorf("marshal result error: %v", err)
	res.WriteHeader(http.StatusInternalServerError)
}

// HTTPGet http get request
func HTTPGet(httpClient *http.Client, url string, result interface{}, headers ...url.Values) (ret []byte, err error) {
	request, requestErr := http.NewRequest("GET", url, nil)
	if requestErr != nil {
		err = requestErr
		log.Errorf("construct request failed, url:%s, err:%s", url, err.Error())
		return
	}

	for _, val := range headers {
		for k, v := range val {
			request.Header.Set(k, v[0])
		}
	}

	response, responseErr := httpClient.Do(request)
	if responseErr != nil {
		err = responseErr
		log.Errorf("get request failed, err:%s", err.Error())
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		err = fmt.Errorf("unexpect statusCode, statusCode:%d", response.StatusCode)
		return
	}

	content, contentErr := io.ReadAll(response.Body)
	if contentErr != nil {
		err = contentErr
		log.Errorf("read respose data failed, err:%s", err.Error())
		return
	}

	if result != nil {
		err = json.Unmarshal(content, result)
		if err != nil {
			log.Errorf("unmarshal data failed, err:%s", err.Error())
			return
		}
	}

	ret = content
	return
}

// HTTPPost http post request
func HTTPPost(httpClient *http.Client, url string, param interface{}, result interface{}, headers ...url.Values) (ret []byte, err error) {
	byteBuff := bytes.NewBuffer(nil)
	if param != nil {
		data, dataErr := json.Marshal(param)
		if dataErr != nil {
			err = dataErr
			log.Errorf("marshal param failed, err:%s", err.Error())
			return
		}

		byteBuff.Write(data)
	}

	request, requestErr := http.NewRequest("POST", url, byteBuff)
	if requestErr != nil {
		err = requestErr
		log.Errorf("construct request failed, url:%s, err:%s", url, err.Error())
		return
	}

	request.Header.Set("content-type", "application/json")
	for _, val := range headers {
		for k, v := range val {
			request.Header.Set(k, v[0])
		}
	}

	response, responseErr := httpClient.Do(request)
	if responseErr != nil {
		err = responseErr
		log.Errorf("post request failed, err:%s", err.Error())
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		err = fmt.Errorf("unexpect statusCode, statusCode:%d", response.StatusCode)
		return
	}

	content, contentErr := io.ReadAll(response.Body)
	if contentErr != nil {
		err = contentErr
		log.Errorf("read respose data failed, err:%s", err.Error())
		return
	}

	if result != nil {
		err = json.Unmarshal(content, result)
		if err != nil {
			log.Errorf("unmarshal data failed, err:%s", err.Error())
			return
		}
	}

	ret = content
	return
}

// HTTPPut http put request
func HTTPPut(httpClient *http.Client, url string, param interface{}, result interface{}, headers ...url.Values) (ret []byte, err error) {
	byteBuff := bytes.NewBuffer(nil)
	if param != nil {
		data, dataErr := json.Marshal(param)
		if dataErr != nil {
			err = dataErr
			log.Errorf("marshal param failed, err:%s", err.Error())
			return
		}

		byteBuff.Write(data)
	}

	request, requestErr := http.NewRequest("PUT", url, byteBuff)
	if requestErr != nil {
		err = requestErr
		log.Errorf("construct request failed, url:%s, err:%s", url, err.Error())
		return
	}

	request.Header.Set("content-type", "application/json")
	for _, val := range headers {
		for k, v := range val {
			request.Header.Set(k, v[0])
		}
	}
	response, responseErr := httpClient.Do(request)
	if responseErr != nil {
		err = responseErr
		log.Errorf("put request failed, err:%s", err.Error())
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		err = fmt.Errorf("unexpect statusCode, statusCode:%d", response.StatusCode)
		return
	}

	content, contentErr := io.ReadAll(response.Body)
	if contentErr != nil {
		err = contentErr
		log.Errorf("read respose data failed, err:%s", err.Error())
		return
	}

	if result != nil {
		err = json.Unmarshal(content, result)
		if err != nil {
			log.Errorf("unmarshal data failed, err:%s", err.Error())
			return
		}
	}

	ret = content
	return
}

// HTTPDelete http delete request
func HTTPDelete(httpClient *http.Client, url string, result interface{}, headers ...url.Values) (ret []byte, err error) {
	request, requestErr := http.NewRequest("DELETE", url, nil)
	if requestErr != nil {
		err = requestErr
		log.Errorf("construct request failed, url:%s, err:%s", url, err.Error())
		return
	}
	for _, val := range headers {
		for k, v := range val {
			request.Header.Set(k, v[0])
		}
	}

	response, responseErr := httpClient.Do(request)
	if responseErr != nil {
		err = responseErr
		log.Errorf("delete request failed, err:%s", err.Error())
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		err = fmt.Errorf("unexpect statusCode, statusCode:%d", response.StatusCode)
		return
	}

	content, contentErr := io.ReadAll(response.Body)
	if contentErr != nil {
		err = contentErr
		log.Errorf("read respose data failed, err:%s", err.Error())
		return
	}

	if result != nil {
		err = json.Unmarshal(content, result)
		if err != nil {
			log.Errorf("unmarshal data failed, err:%s", err.Error())
			return
		}
	}

	ret = content
	return
}

// HTTPDownload http download file
func HTTPDownload(httpClient *http.Client, url string, filePath string, headers ...url.Values) (string, error) {
	request, requestErr := http.NewRequest("GET", url, nil)
	if requestErr != nil {
		log.Errorf("construct request failed, url:%s, err:%s", url, requestErr.Error())
		return "", requestErr
	}

	for _, val := range headers {
		for k, v := range val {
			request.Header.Set(k, v[0])
		}
	}

	response, responseErr := httpClient.Do(request)
	if responseErr != nil {
		log.Errorf("get request failed, err:%s", responseErr.Error())
		return "", responseErr
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("unexpect statusCode, statusCode:%d", response.StatusCode)
		return "", errors.New(msg)
	}

	f, err := os.OpenFile(filePath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Errorf("open destination file failed, err:%s", err.Error())
		return "", err
	}
	defer f.Close()

	_, err = io.Copy(f, response.Body)
	if err != nil {
		log.Errorf("write destination file content exception, err:%s", err.Error())
		return "", err
	}

	return filePath, nil
}

// HTTPUpload http upload file
func HTTPUpload(httpClient *http.Client, url, fileItem, filePath string, result interface{}, headers ...url.Values) error {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	//关键的一步操作
	fileWriter, err := bodyWriter.CreateFormFile(fileItem, filePath)
	if err != nil {
		log.Errorf("error writing to buffer")
		return err
	}

	//打开文件句柄操作
	fh, err := os.Open(filePath)
	if err != nil {
		log.Errorf("error opening file")
		return err
	}
	defer fh.Close()

	//iocopy
	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		return err
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	request, requestErr := http.NewRequest("POST", url, bodyBuf)
	if requestErr != nil {
		err = requestErr
		log.Errorf("construct request failed, url:%s, err:%s", url, err.Error())
		return err
	}

	request.Header.Set("content-type", contentType)
	for _, val := range headers {
		for k, v := range val {
			request.Header.Set(k, v[0])
		}
	}

	response, responseErr := httpClient.Do(request)
	if responseErr != nil {
		log.Errorf("post request failed, err:%s", responseErr.Error())
		return responseErr
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("unexpect statusCode, statusCode:%d", response.StatusCode)
		return errors.New(msg)
	}

	if result != nil {
		content, err := io.ReadAll(response.Body)
		if err != nil {
			log.Errorf("read respose data failed, err:%s", err.Error())
			return err
		}

		err = json.Unmarshal(content, result)
		if err != nil {
			log.Errorf("unmarshal data failed, err:%s", err.Error())
			return err
		}
	}

	return nil
}

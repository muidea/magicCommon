package net

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/muidea/magicCommon/foundation/util"
)

type maxBytesReader struct {
	res http.ResponseWriter
	req io.ReadCloser // underlying reader
	n   int64         // max bytes remaining
	err error         // sticky error
}

func (l *maxBytesReader) tooLarge() (n int, err error) {
	l.err = errors.New("http: request body too large")
	return 0, l.err
}

func (l *maxBytesReader) Read(p []byte) (n int, err error) {
	if l.err != nil {
		return 0, l.err
	}
	if len(p) == 0 {
		return 0, nil
	}
	// If they asked for a 32KB byte read but only 5 bytes are
	// remaining, no need to read 32KB. 6 bytes will answer the
	// question of the whether we hit the limit or go past it.
	if int64(len(p)) > l.n+1 {
		p = p[:l.n+1]
	}
	n, err = l.req.Read(p)

	if int64(n) <= l.n {
		l.n -= int64(n)
		l.err = err
		return n, err
	}

	n = int(l.n)
	l.n = 0

	// The server code and client code both use
	// maxBytesReader. This "requestTooLarge" check is
	// only used by the server code. To prevent binaries
	// which only using the HTTP Client code (such as
	// cmd/go) from also linking in the HTTP server, don't
	// use a static type assertion to the server
	// "*response" type. Check this interface instead:
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
	return l.req.Close()
}

// GetHTTPRequestBody get http request body
func GetHTTPRequestBody(req *http.Request) (ret []byte, err error) {
	var reader io.Reader = req.Body
	maxFormSize := int64(1<<63 - 1)
	if _, ok := req.Body.(*maxBytesReader); !ok {
		maxFormSize = int64(10 << 20) // 10 MB is a lot of text.
		reader = io.LimitReader(req.Body, maxFormSize+1)
	}

	payload, payloadErr := ioutil.ReadAll(reader)
	if payloadErr != nil {
		err = payloadErr
		return
	}

	if int64(len(payload)) > maxFormSize {
		err = errors.New("http: request body too large")
		return
	}

	ret = payload
	return
}

// ParseJSONBody 解析http body请求提交的json数据
func ParseJSONBody(req *http.Request, param interface{}) error {
	util.ValidataPtr(param)

	if req.Body == nil {
		return errors.New("missing form body")
	}

	contentType, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil {
		return err
	}

	switch {
	case contentType == "application/json":
		payload, payloadErr := GetHTTPRequestBody(req)
		if payloadErr != nil {
			return payloadErr
		}

		err = json.Unmarshal(payload, param)
		if err != nil {
			return err
		}

	default:
		return errors.New("invalid contentType, contentType:" + contentType)
	}

	return nil
}

// HTTPGet http get request
func HTTPGet(httpClient *http.Client, url string, result interface{}) (ret []byte, err error) {
	response, responseErr := httpClient.Get(url)
	if responseErr != nil {
		err = responseErr
		log.Printf("get request failed, err:%s", err.Error())
		return
	}

	if response.StatusCode != http.StatusOK {
		err = fmt.Errorf("unexpect statusCode, statusCode:%d", response.StatusCode)
		return
	}

	content, contentErr := ioutil.ReadAll(response.Body)
	if contentErr != nil {
		err = contentErr
		log.Printf("read respose data failed, err:%s", err.Error())
		return
	}

	if result != nil {
		err = json.Unmarshal(content, result)
		if err != nil {
			log.Printf("unmarshal data failed, err:%s", err.Error())
			return
		}
	}

	ret = content
	return
}

// HTTPPost http post request
func HTTPPost(httpClient *http.Client, url string, param interface{}, result interface{}) (ret []byte, err error) {
	var bufferReader *bytes.Buffer
	if param != nil {
		data, dataErr := json.Marshal(param)
		if dataErr != nil {
			err = dataErr
			log.Printf("marshal param failed, err:%s", err.Error())
			return
		}

		bufferReader = bytes.NewBuffer(data)
	}

	request, requestErr := http.NewRequest("POST", url, bufferReader)
	if requestErr != nil {
		err = requestErr
		log.Printf("construct request failed, url:%s, err:%s", url, err.Error())
		return
	}

	request.Header.Set("content-type", "application/json")
	response, responseErr := httpClient.Do(request)
	if responseErr != nil {
		err = responseErr
		log.Printf("post request failed, err:%s", err.Error())
		return
	}

	if response.StatusCode != http.StatusOK {
		err = fmt.Errorf("unexpect statusCode, statusCode:%d", response.StatusCode)
		return
	}

	content, contentErr := ioutil.ReadAll(response.Body)
	if contentErr != nil {
		err = contentErr
		log.Printf("read respose data failed, err:%s", err.Error())
		return
	}

	if result != nil {
		err = json.Unmarshal(content, result)
		if err != nil {
			log.Printf("unmarshal data failed, err:%s", err.Error())
			return
		}
	}

	ret = content
	return
}

// HTTPPut http post request
func HTTPPut(httpClient *http.Client, url string, param interface{}, result interface{}) (ret []byte, err error) {
	var bufferReader *bytes.Buffer
	if param != nil {
		data, dataErr := json.Marshal(param)
		if dataErr != nil {
			err = dataErr
			log.Printf("marshal param failed, err:%s", err.Error())
			return
		}

		bufferReader = bytes.NewBuffer(data)
	}

	request, requestErr := http.NewRequest("PUT", url, bufferReader)
	if requestErr != nil {
		err = requestErr
		log.Printf("construct request failed, url:%s, err:%s", url, err.Error())
		return
	}

	request.Header.Set("content-type", "application/json")
	response, responseErr := httpClient.Do(request)
	if responseErr != nil {
		err = responseErr
		log.Printf("post request failed, err:%s", err.Error())
		return
	}

	if response.StatusCode != http.StatusOK {
		err = fmt.Errorf("unexpect statusCode, statusCode:%d", response.StatusCode)
		return
	}

	content, contentErr := ioutil.ReadAll(response.Body)
	if contentErr != nil {
		err = contentErr
		log.Printf("read respose data failed, err:%s", err.Error())
		return
	}

	if result != nil {
		err = json.Unmarshal(content, result)
		if err != nil {
			log.Printf("unmarshal data failed, err:%s", err.Error())
			return
		}
	}

	ret = content
	return
}

// HTTPDelete http delete request
func HTTPDelete(httpClient *http.Client, url string, param interface{}, result interface{}) (ret []byte, err error) {
	var bufferReader *bytes.Buffer
	if param != nil {
		data, dataErr := json.Marshal(param)
		if dataErr != nil {
			err = dataErr
			log.Printf("marshal param failed, err:%s", err.Error())
			return
		}

		bufferReader = bytes.NewBuffer(data)
	}

	request, requestErr := http.NewRequest("DELETE", url, bufferReader)
	if requestErr != nil {
		err = requestErr
		log.Printf("construct request failed, url:%s, err:%s", url, err.Error())
		return
	}

	response, responseErr := httpClient.Do(request)
	if responseErr != nil {
		err = responseErr
		log.Printf("post request failed, err:%s", err.Error())
		return
	}

	if response.StatusCode != http.StatusOK {
		err = fmt.Errorf("unexpect statusCode, statusCode:%d", response.StatusCode)
		return
	}

	content, contentErr := ioutil.ReadAll(response.Body)
	if contentErr != nil {
		err = contentErr
		log.Printf("read respose data failed, err:%s", err.Error())
		return
	}

	if result != nil {
		err = json.Unmarshal(content, result)
		if err != nil {
			log.Printf("unmarshal data failed, err:%s", err.Error())
			return
		}
	}

	ret = content
	return
}

// HTTPDownload http download file
func HTTPDownload(httpClient *http.Client, url string, filePath string) (string, error) {
	response, err := httpClient.Get(url)
	if err != nil {
		log.Printf("get request failed, err:%s", err.Error())
		return "", err
	}

	if response.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("unexpect statusCode, statusCode:%d", response.StatusCode)
		return "", errors.New(msg)
	}

	f, err := os.OpenFile(filePath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Printf("open destination file failed, err:%s", err.Error())
		return "", err
	}
	defer f.Close()

	_, err = io.Copy(f, response.Body)
	if err != nil {
		log.Printf("write destination file content exception, err:%s", err.Error())
		return "", err
	}

	return filePath, nil
}

// HTTPUpload http upload file
func HTTPUpload(httpClient *http.Client, url, fileItem, filePath string, result interface{}) error {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	//关键的一步操作
	fileWriter, err := bodyWriter.CreateFormFile(fileItem, filePath)
	if err != nil {
		fmt.Println("error writing to buffer")
		return err
	}

	//打开文件句柄操作
	fh, err := os.Open(filePath)
	if err != nil {
		fmt.Println("error opening file")
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

	response, err := http.Post(url, contentType, bodyBuf)
	if response.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("unexpect statusCode, statusCode:%d", response.StatusCode)
		return errors.New(msg)
	}

	if result != nil {
		content, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Printf("read respose data failed, err:%s", err.Error())
			return err
		}

		err = json.Unmarshal(content, result)
		if err != nil {
			log.Printf("unmarshal data failed, err:%s", err.Error())
			return err
		}
	}

	return nil
}

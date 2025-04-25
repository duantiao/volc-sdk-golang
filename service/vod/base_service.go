package vod

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"hash/crc64"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/avast/retry-go"
	"github.com/volcengine/volc-sdk-golang/base"
	model_base "github.com/volcengine/volc-sdk-golang/service/base/models/base"
	"github.com/volcengine/volc-sdk-golang/service/vod/models/business"
	"github.com/volcengine/volc-sdk-golang/service/vod/models/request"
	"github.com/volcengine/volc-sdk-golang/service/vod/models/response"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/volcengine/volc-sdk-golang/service/vod/upload/consts"
	"github.com/volcengine/volc-sdk-golang/service/vod/upload/model"
)

func (p *Vod) GetSubtitleAuthToken(req *request.VodGetSubtitleInfoListRequest, tokenExpireTime int) (string, error) {
	if len(req.GetVid()) == 0 {
		return "", errors.New("传入的Vid为空")
	}
	query := url.Values{
		"Vid":    []string{req.GetVid()},
		"Status": []string{"Published"},
	}

	if tokenExpireTime > 0 {
		query.Add("X-Expires", strconv.Itoa(tokenExpireTime))
	}

	if getSubtitleInfoAuthToken, err := p.GetSignUrl("GetSubtitleInfoList", query); err == nil {
		ret := map[string]string{}
		ret["GetSubtitleAuthToken"] = getSubtitleInfoAuthToken
		b, err := json.Marshal(ret)
		if err != nil {
			return "", err
		}
		return base64.StdEncoding.EncodeToString(b), nil
	} else {
		return "", err
	}
}

func (p *Vod) GetPrivateDrmAuthToken(req *request.VodGetPrivateDrmPlayAuthRequest, tokenExpireTime int) (string, error) {
	if len(req.GetVid()) == 0 {
		return "", errors.New("传入的Vid为空")
	}
	query := url.Values{
		"Vid": []string{req.GetVid()},
	}

	if len(req.GetPlayAuthIds()) > 0 {
		query.Add("PlayAuthIds", req.GetPlayAuthIds())
	}
	if len(req.GetDrmType()) > 0 {
		query.Add("DrmType", req.GetDrmType())
		switch req.GetDrmType() {
		case "appdevice", "webdevice":
			if len(req.GetUnionInfo()) == 0 {
				return "", errors.New("invalid unionInfo")
			}
		}
	}
	if len(req.GetUnionInfo()) > 0 {
		query.Add("UnionInfo", req.GetUnionInfo())
	}
	if tokenExpireTime > 0 {
		query.Add("X-Expires", strconv.Itoa(tokenExpireTime))
	}

	if getPrivateDrmAuthToken, err := p.GetSignUrl("GetPrivateDrmPlayAuth", query); err == nil {
		return getPrivateDrmAuthToken, nil
	} else {
		return "", err
	}
}

func (p *Vod) CreateSha1HlsDrmAuthToken(expireSeconds int64) (auth string, err error) {
	return p.createHlsDrmAuthToken(DSAHmacSha1, expireSeconds)
}

func (p *Vod) createHlsDrmAuthToken(authAlgorithm string, expireSeconds int64) (string, error) {
	if expireSeconds == 0 {
		return "", errors.New("invalid expire")
	}

	token, err := createAuth(authAlgorithm, Version2, p.ServiceInfo.Credentials.AccessKeyID,
		p.ServiceInfo.Credentials.SecretAccessKey, p.ServiceInfo.Credentials.Region, expireSeconds)
	if err != nil {
		return "", err
	}

	query := url.Values{}
	query.Set("DrmAuthToken", token)
	query.Set("X-Expires", strconv.FormatInt(expireSeconds, 10))
	if getAuth, err := p.GetSignUrl("GetHlsDecryptionKey", query); err == nil {
		return getAuth, nil
	} else {
		return "", err
	}
}

func (p *Vod) GetPlayAuthToken(req *request.VodGetPlayInfoRequest, tokenExpireTime int) (string, error) {
	if len(req.GetVid()) == 0 {
		return "", errors.New("传入的Vid为空")
	}
	query := url.Values{}
	marshaler := protojson.MarshalOptions{
		Multiline:       false,
		AllowPartial:    false,
		UseProtoNames:   true,
		UseEnumNumbers:  false,
		EmitUnpopulated: false,
	}
	jsonData := marshaler.Format(req)
	reqMap := map[string]interface{}{}
	err := json.Unmarshal([]byte(jsonData), &reqMap)
	if err != nil {
		return "", err
	}
	for k, v := range reqMap {
		var sv string
		switch ov := v.(type) {
		case string:
			sv = ov
		case int:
			sv = strconv.FormatInt(int64(ov), 10)
		case uint:
			sv = strconv.FormatUint(uint64(ov), 10)
		case int8:
			sv = strconv.FormatInt(int64(ov), 10)
		case uint8:
			sv = strconv.FormatUint(uint64(ov), 10)
		case int16:
			sv = strconv.FormatInt(int64(ov), 10)
		case uint16:
			sv = strconv.FormatUint(uint64(ov), 10)
		case int32:
			sv = strconv.FormatInt(int64(ov), 10)
		case uint32:
			sv = strconv.FormatUint(uint64(ov), 10)
		case int64:
			sv = strconv.FormatInt(ov, 10)
		case uint64:
			sv = strconv.FormatUint(ov, 10)
		case bool:
			sv = strconv.FormatBool(ov)
		case float32:
			sv = strconv.FormatFloat(float64(ov), 'f', -1, 32)
		case float64:
			sv = strconv.FormatFloat(ov, 'f', -1, 64)
		case []byte:
			sv = string(ov)
		default:
			v2, e2 := json.Marshal(ov)
			if e2 != nil {
				return "", e2
			}
			sv = string(v2)
		}
		query.Set(k, sv)
	}
	if tokenExpireTime > 0 {
		query.Add("X-Expires", strconv.Itoa(tokenExpireTime))
	}
	if getPlayInfoToken, err := p.GetSignUrl("GetPlayInfo", query); err == nil {
		ret := map[string]string{}
		ret["GetPlayInfoToken"] = getPlayInfoToken
		ret["TokenVersion"] = "V2"
		b, err := json.Marshal(ret)
		if err != nil {
			return "", err
		}
		return base64.StdEncoding.EncodeToString(b), nil
	} else {
		return "", err
	}
}

func (p *Vod) UploadObjectWithCallback(filePath string, spaceName string, callbackArgs string, fileName, fileExtension string, funcs string) (*response.VodCommitUploadInfoResponse, int, error) {
	file, err := os.Open(filepath.Clean(filePath))
	if err != nil {
		return nil, -1, err
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		return nil, -1, err
	}

	req := &model.VodUploadMediaInnerFuncRequest{
		FilePath:      filePath,
		Rd:            file,
		Size:          stat.Size(),
		SpaceName:     spaceName,
		FileType:      "object",
		CallbackArgs:  callbackArgs,
		Funcs:         funcs,
		FileName:      fileName,
		FileExtension: fileExtension,
	}
	return p.UploadMediaInner(req)
}

func (p *Vod) UploadObjectWithCallbackV2(uploadReq *request.VodUploadObjectRequest) (*response.VodCommitUploadInfoResponse, int, error) {
	file, err := os.Open(filepath.Clean(uploadReq.FilePath))
	if err != nil {
		return nil, -1, err
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		return nil, -1, err
	}

	req := &model.VodUploadMediaInnerFuncRequest{
		FilePath:          uploadReq.FilePath,
		Rd:                file,
		Size:              stat.Size(),
		SpaceName:         uploadReq.SpaceName,
		FileType:          "object",
		CallbackArgs:      uploadReq.CallbackArgs,
		Funcs:             uploadReq.Functions,
		FileName:          uploadReq.FileName,
		FileExtension:     uploadReq.FileExtension,
		ClientIDCMode:     uploadReq.ClientIDCMode,
		ClientNetWorkMode: uploadReq.ClientNetWorkMode,
		ChunkSize:         uploadReq.ChunkSize,
	}
	return p.UploadMediaInner(req)
}

func (p *Vod) UploadMediaWithCallback(mediaRequset *request.VodUploadMediaRequest) (*response.VodCommitUploadInfoResponse, int, error) {
	file, err := os.Open(filepath.Clean(mediaRequset.GetFilePath()))
	if err != nil {
		return nil, -1, err
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		return nil, -1, err
	}

	req := &model.VodUploadMediaInnerFuncRequest{
		FilePath:          mediaRequset.GetFilePath(),
		Rd:                file,
		Size:              stat.Size(),
		ParallelNum:       int(mediaRequset.GetParallelNum()),
		SpaceName:         mediaRequset.GetSpaceName(),
		CallbackArgs:      mediaRequset.GetCallbackArgs(),
		Funcs:             mediaRequset.GetFunctions(),
		FileName:          mediaRequset.GetFileName(),
		FileExtension:     mediaRequset.GetFileExtension(),
		VodUploadSource:   mediaRequset.GetVodUploadSource(),
		StorageClass:      mediaRequset.StorageClass,
		ClientNetWorkMode: mediaRequset.ClientNetWorkMode,
		ClientIDCMode:     mediaRequset.ClientIDCMode,
		ExpireTime:        mediaRequset.ExpireTime,
		UploadHostPrefer:  mediaRequset.UploadHostPrefer,
		ChunkSize:         mediaRequset.ChunkSize,
	}
	return p.UploadMediaInner(req)
}

func (p *Vod) UploadMaterialWithCallback(materialRequest *request.VodUploadMaterialRequest) (*response.VodCommitUploadInfoResponse, int, error) {
	file, err := os.Open(filepath.Clean(materialRequest.GetFilePath()))
	if err != nil {
		return nil, -1, err
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		return nil, -1, err
	}

	req := &model.VodUploadMediaInnerFuncRequest{
		FilePath:          materialRequest.GetFilePath(),
		Rd:                file,
		Size:              stat.Size(),
		ParallelNum:       int(materialRequest.GetParallelNum()),
		SpaceName:         materialRequest.GetSpaceName(),
		FileType:          materialRequest.GetFileType(),
		CallbackArgs:      materialRequest.GetCallbackArgs(),
		Funcs:             materialRequest.GetFunctions(),
		FileName:          materialRequest.GetFileName(),
		FileExtension:     materialRequest.GetFileExtension(),
		UploadHostPrefer:  materialRequest.GetUploadHostPrefer(),
		ChunkSize:         materialRequest.ChunkSize,
		ClientNetWorkMode: materialRequest.GetClientNetWorkMode(),
		ClientIDCMode:     materialRequest.GetClientIDCMode(),
	}
	return p.UploadMediaInner(req)
}

func (p *Vod) UploadMediaInner(uploadMediaInnerRequest *model.VodUploadMediaInnerFuncRequest) (r *response.VodCommitUploadInfoResponse, c int, e error) {
	now := time.Now()
	defer func() {
		var (
			msg  string
			code = 200
		)
		if e != nil {
			msg = e.Error()
			code = 500
		}
		reporter.report(p, p.buildDefaultUploadReport(uploadMediaInnerRequest.SpaceName, time.Since(now).Microseconds(), code, 0, "", actionFinishUpload, "", msg))
	}()
	req := &model.VodUploadFuncRequest{
		FilePath:          uploadMediaInnerRequest.FilePath,
		Rd:                uploadMediaInnerRequest.Rd,
		Size:              uploadMediaInnerRequest.Size,
		ParallelNum:       uploadMediaInnerRequest.ParallelNum,
		SpaceName:         uploadMediaInnerRequest.SpaceName,
		FileType:          uploadMediaInnerRequest.FileType,
		FileName:          uploadMediaInnerRequest.FileName,
		FileExtension:     uploadMediaInnerRequest.FileExtension,
		StorageClass:      uploadMediaInnerRequest.StorageClass,
		ClientNetWorkMode: uploadMediaInnerRequest.ClientNetWorkMode,
		ClientIDCMode:     uploadMediaInnerRequest.ClientIDCMode,
		UploadHostPrefer:  uploadMediaInnerRequest.UploadHostPrefer,
		ChunkSize:         uploadMediaInnerRequest.ChunkSize,
	}
	logId, sessionKey, err, code := p.Upload(req)
	if err != nil {
		return p.fillCommitUploadInfoResponseWhenError(logId, err.Error()), code, err
	}

	commitRequest := &request.VodCommitUploadInfoRequest{
		SpaceName:       uploadMediaInnerRequest.SpaceName,
		SessionKey:      sessionKey,
		CallbackArgs:    uploadMediaInnerRequest.CallbackArgs,
		Functions:       uploadMediaInnerRequest.Funcs,
		VodUploadSource: uploadMediaInnerRequest.VodUploadSource,
		ExpireTime:      uploadMediaInnerRequest.ExpireTime,
	}

	now1 := time.Now()
	commitResp, code, err := p.CommitUploadInfo(commitRequest)
	if err != nil {
		if commitResp == nil {
			reporter.report(p, p.buildDefaultUploadReport(uploadMediaInnerRequest.SpaceName, time.Since(now1).Microseconds(), code, 0, "", actionCommitUploadInfo, "", err.Error()))
		}
		return commitResp, code, err
	}
	return commitResp, code, nil
}

func WithUploadKeyPtn(ptn string) model.UploadAuthOpt {
	return func(o *model.UploadAuthOption) {
		o.KeyPtn = ptn
	}
}

func WithUploadSpaceNames(spaceNames []string) model.UploadAuthOpt {
	return func(op *model.UploadAuthOption) {
		op.SpaceNames = spaceNames
	}
}

func WithUploadPolicy(policy *model.UploadPolicy) model.UploadAuthOpt {
	return func(op *model.UploadAuthOption) {
		op.UploadPolicy = policy
	}
}

func (p *Vod) GetUploadAuthWithExpiredTime(expiredTime time.Duration, opt ...model.UploadAuthOpt) (*base.SecurityToken2, error) {
	inlinePolicy := new(base.Policy)
	op := &model.UploadAuthOption{}
	for _, o := range opt {
		o(op)
	}

	spaceRes := make([]string, 0)
	if len(op.SpaceNames) != 0 {
		for _, space := range op.SpaceNames {
			spaceRes = append(spaceRes, fmt.Sprintf(consts.ResourceSpaceNameTRN, space))
		}
	}

	resources := make([]string, 0)
	inlinePolicy.Statement = append(inlinePolicy.Statement, base.NewAllowStatement([]string{"vod:ApplyUploadInfo"}, spaceRes))
	inlinePolicy.Statement = append(inlinePolicy.Statement, base.NewAllowStatement([]string{"vod:CommitUploadInfo"}, resources))

	if op.KeyPtn != "" {
		inlinePolicy.Statement = append(inlinePolicy.Statement, base.NewAllowStatement([]string{"FileNamePtn"}, []string{op.KeyPtn}))
	}

	if op.UploadPolicy != nil {
		policyStr, err := json.Marshal(op.UploadPolicy)
		if err != nil {
			return nil, err
		}
		inlinePolicy.Statement = append(inlinePolicy.Statement, base.NewAllowStatement([]string{"UploadPolicy"}, []string{string(policyStr)}))
	}
	return p.SignSts2(inlinePolicy, expiredTime)
}

func (p *Vod) GetUploadAuth(opt ...model.UploadAuthOpt) (*base.SecurityToken2, error) {
	return p.GetUploadAuthWithExpiredTime(time.Hour, opt...)
}

func (p *Vod) fillCommitUploadInfoResponseWhenError(logId, errMsg string) *response.VodCommitUploadInfoResponse {
	commitUploadInfoRespone := &response.VodCommitUploadInfoResponse{
		ResponseMetadata: &model_base.ResponseMetadata{
			RequestId: logId,
			Service:   "vod",
			Error:     &model_base.ResponseError{Message: errMsg},
		},
	}
	return commitUploadInfoRespone
}

func (p *Vod) Upload(vodUploadFuncRequest *model.VodUploadFuncRequest) (string, string, error, int) {
	if vodUploadFuncRequest.Size == 0 {
		return "", "", fmt.Errorf("file size is zero"), http.StatusBadRequest
	}
	if vodUploadFuncRequest.ChunkSize < consts.MinChunckSize {
		vodUploadFuncRequest.ChunkSize = consts.MinChunckSize
	}

	applyRequest := &request.VodApplyUploadInfoRequest{
		SpaceName:         vodUploadFuncRequest.SpaceName,
		FileType:          vodUploadFuncRequest.FileType,
		FileName:          vodUploadFuncRequest.FileName,
		FileExtension:     vodUploadFuncRequest.FileExtension,
		StorageClass:      vodUploadFuncRequest.StorageClass,
		ClientNetWorkMode: vodUploadFuncRequest.ClientNetWorkMode,
		ClientIDCMode:     vodUploadFuncRequest.ClientIDCMode,
		NeedFallback:      true, // default set
		UploadHostPrefer:  vodUploadFuncRequest.UploadHostPrefer,
		FileSize:          float64(vodUploadFuncRequest.Size),
	}

	now := time.Now()
	resp, code, err := p.ApplyUploadInfo(applyRequest)
	logId := resp.GetResponseMetadata().GetRequestId()
	if err != nil {
		if resp == nil {
			reporter.report(p, p.buildDefaultUploadReport(vodUploadFuncRequest.SpaceName, time.Since(now).Microseconds(), code, 0, logId, actionApplyUploadInfo, "", err.Error()))
		}
		return logId, "", err, code
	}

	if resp.ResponseMetadata.Error != nil && resp.ResponseMetadata.Error.Code != "0" {
		return logId, "", fmt.Errorf("%+v", resp.ResponseMetadata.Error), code
	}

	// vpc upload
	if vpcUploadAddress := resp.GetResult().GetData().GetVpcTosUploadAddress(); vpcUploadAddress != nil {
		err := p.vpcUpload(vpcUploadAddress, vodUploadFuncRequest)
		if err != nil {
			return logId, "", err, http.StatusBadRequest
		}
		uploadAddress := resp.GetResult().GetData().GetUploadAddress()
		oid := uploadAddress.StoreInfos[0].GetStoreUri()
		sessionKey := uploadAddress.GetSessionKey()
		return oid, sessionKey, nil, http.StatusOK
	}

	// using candidate address first
	candidateUploadAddress := resp.GetResult().GetData().GetCandidateUploadAddresses()
	var allUploadAddress []*business.UploadAddress
	if candidateUploadAddress != nil {
		allUploadAddress = append(allUploadAddress, candidateUploadAddress.GetMainUploadAddresses()...)
		allUploadAddress = append(allUploadAddress, candidateUploadAddress.GetBackupUploadAddresses()...)
		allUploadAddress = append(allUploadAddress, candidateUploadAddress.GetFallbackUploadAddresses()...)
	}
	if len(allUploadAddress) > 0 {
		client := &http.Client{}
		var bts []byte
		for i, uploadAddress := range allUploadAddress {
			if len(uploadAddress.GetUploadHosts()) == 0 || len(uploadAddress.GetStoreInfos()) == 0 || uploadAddress.GetStoreInfos()[0] == nil {
				continue
			}
			tosHost := uploadAddress.GetUploadHosts()[0]
			oid := uploadAddress.StoreInfos[0].GetStoreUri()
			sessionKey := uploadAddress.GetSessionKey()
			auth := uploadAddress.GetStoreInfos()[0].GetAuth()

			var lazyUploadFn func() error
			if vodUploadFuncRequest.ParallelNum == 0 {
				vodUploadFuncRequest.ParallelNum = 1
			}
			uploadPart := model.UploadPartCommon{
				Client:       client,
				TosHost:      tosHost,
				Oid:          oid,
				Auth:         auth,
				SpaceName:    vodUploadFuncRequest.SpaceName,
				ChunkSize:    vodUploadFuncRequest.ChunkSize,
				FileSize:     vodUploadFuncRequest.Size,
				ParallelNum:  vodUploadFuncRequest.ParallelNum,
				StorageClass: vodUploadFuncRequest.StorageClass,
			}
			if vodUploadFuncRequest.Size < vodUploadFuncRequest.ChunkSize {
				if len(bts) == 0 {
					bts, err = ioutil.ReadAll(vodUploadFuncRequest.Rd)
					if err != nil {
						return logId, "", err, http.StatusBadRequest
					}
				}
				lazyUploadFn = func() error {
					return p.directUpload(bts, uploadPart)
				}
			} else {
				lazyUploadFn = func() error {
					return p.chunkUpload(vodUploadFuncRequest.FilePath, uploadPart)
				}
			}

			retryCount := 1
			// retry 3 times when received specific error code from transporter
			if err := retry.Do(func() error {
				fmt.Printf("using %d host, try %d times\n", i+1, retryCount)
				uploadPart.RetryTimes = retryCount
				retryCount++
				return lazyUploadFn()
			}, retry.RetryIf(func(err error) bool {
				if e, ok := err.(UploadError); ok {
					return e.ErrorCode >= 5000 || e.ErrorCode == 0 && e.Code >= 500
				}
				return false
			}), retry.Attempts(3), retry.LastErrorOnly(true)); err != nil {
				if e, ok := err.(UploadError); ok {
					// next domain
					if !(e.ErrorCode >= 5000 || e.ErrorCode == 0 && e.Code >= 500) {
						return logId, "", err, http.StatusBadRequest
					}
				}
				continue
			}
			return oid, sessionKey, nil, http.StatusOK
		}
		return logId, "", fmt.Errorf("upload failed"), http.StatusBadRequest
	} else {
		uploadAddress := resp.GetResult().GetData().GetUploadAddress()
		if uploadAddress != nil {
			if len(uploadAddress.GetUploadHosts()) == 0 {
				return logId, "", fmt.Errorf("no tos host found"), http.StatusBadRequest
			}
			if len(uploadAddress.GetStoreInfos()) == 0 || (uploadAddress.GetStoreInfos()[0] == nil) {
				return logId, "", fmt.Errorf("no store info found"), http.StatusBadRequest
			}

			tosHost := uploadAddress.GetUploadHosts()[0]
			oid := uploadAddress.StoreInfos[0].GetStoreUri()
			sessionKey := uploadAddress.GetSessionKey()
			auth := uploadAddress.GetStoreInfos()[0].GetAuth()
			if vodUploadFuncRequest.ParallelNum == 0 {
				vodUploadFuncRequest.ParallelNum = 1
			}
			param := model.UploadPartCommon{
				Client:       &http.Client{},
				TosHost:      tosHost,
				Oid:          oid,
				Auth:         auth,
				ChunkSize:    vodUploadFuncRequest.ChunkSize,
				FileSize:     vodUploadFuncRequest.Size,
				ParallelNum:  vodUploadFuncRequest.ParallelNum,
				StorageClass: vodUploadFuncRequest.StorageClass,
				SpaceName:    vodUploadFuncRequest.SpaceName,
			}
			if vodUploadFuncRequest.Size < vodUploadFuncRequest.ChunkSize {
				bts, err := ioutil.ReadAll(vodUploadFuncRequest.Rd)
				if err != nil {
					return logId, "", err, http.StatusBadRequest
				}
				if err := p.directUpload(bts, param); err != nil {
					return logId, "", err, http.StatusBadRequest
				}
			} else {
				if err := p.chunkUpload(vodUploadFuncRequest.FilePath, param); err != nil {
					return logId, "", err, http.StatusBadRequest
				}
			}
			return oid, sessionKey, nil, http.StatusOK
		}
	}
	return logId, "", errors.New("upload address not exist"), http.StatusBadRequest
}

func (p *Vod) directUpload(fileBytes []byte, param model.UploadPartCommon) error {
	tosHost, oid, auth, client, storageClass := param.TosHost, param.Oid, param.Auth, param.Client, param.StorageClass
	checkSum := fmt.Sprintf("%08x", crc32.ChecksumIEEE(fileBytes))
	url := fmt.Sprintf("https://%s/%s", tosHost, oid)
	req, err := http.NewRequest("PUT", url, bytes.NewReader(fileBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-CRC32", checkSum)
	req.Header.Set("Authorization", auth)

	if storageClass == int32(business.StorageClassType_Archive) {
		req.Header.Set("X-Upload-Storage-Class", "archive")
	}
	if storageClass == int32(business.StorageClassType_IA) {
		req.Header.Set("X-Upload-Storage-Class", "ia")
	}

	now := time.Now()
	rsp, err := client.Do(req)
	if err != nil {
		reporter.report(p, p.buildDefaultUploadReport(param.SpaceName, time.Since(now).Microseconds(), 500, param.RetryTimes, "", actionDirectUpload, tosHost, err.Error()))
		return err
	}
	b, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		reporter.report(p, p.buildDefaultUploadReport(param.SpaceName, time.Since(now).Microseconds(), rsp.StatusCode, param.RetryTimes, rsp.Header.Get(logHeader), actionDirectUpload, tosHost, err.Error()))
		return err
	}
	res := &model.UploadPartCommonResponse{}
	if err := json.Unmarshal(b, res); err != nil {
		errStr := fmt.Sprintf("unmarshal direct upload response failed: %v, got result: %s", err, string(b))
		reporter.report(p, p.buildDefaultUploadReport(param.SpaceName, time.Since(now).Microseconds(), rsp.StatusCode, param.RetryTimes, rsp.Header.Get(logHeader), actionDirectUpload, tosHost, errStr))
		return err
	}
	if res.Success != 0 {
		return errors.New(res.Error.Message)
	}
	return nil
}

func (p *Vod) directUploadStream(content io.Reader, param model.UploadPartCommon) error {
	tosHost, oid, auth, client, storageClass := param.TosHost, param.Oid, param.Auth, param.Client, param.StorageClass
	checkSum := "Ignore"
	url := fmt.Sprintf("https://%s/%s", tosHost, oid)
	req, err := http.NewRequest("PUT", url, content)
	if err != nil {
		return err
	}
	req.Header.Set("Content-CRC32", checkSum)
	req.Header.Set("Authorization", auth)

	if storageClass == int32(business.StorageClassType_Archive) {
		req.Header.Set("X-Upload-Storage-Class", "archive")
	}
	if storageClass == int32(business.StorageClassType_IA) {
		req.Header.Set("X-Upload-Storage-Class", "ia")
	}

	now := time.Now()
	rsp, err := client.Do(req)
	if err != nil {
		reporter.report(p, p.buildDefaultUploadReport(param.SpaceName, time.Since(now).Microseconds(), 500, param.RetryTimes, "", actionDirectUpload, tosHost, err.Error()))
		return err
	}
	b, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		reporter.report(p, p.buildDefaultUploadReport(param.SpaceName, time.Since(now).Microseconds(), rsp.StatusCode, param.RetryTimes, rsp.Header.Get(logHeader), actionDirectUpload, tosHost, err.Error()))
		return err
	}
	res := &model.UploadPartCommonResponse{}
	if err := json.Unmarshal(b, res); err != nil {
		errStr := fmt.Sprintf("unmarshal stream direct upload response failed: %v, got result: %s", err, string(b))
		reporter.report(p, p.buildDefaultUploadReport(param.SpaceName, time.Since(now).Microseconds(), rsp.StatusCode, param.RetryTimes, rsp.Header.Get(logHeader), actionDirectUpload, tosHost, errStr))
		return err
	}
	if res.Success != 0 {
		return errors.New(res.Error.Message)
	}
	return nil
}

func crc64ecma(f io.Reader) (string, error) {
	crc64Hash := crc64.New(crc64.MakeTable(crc64.ECMA))
	if _, err := io.Copy(crc64Hash, f); err != nil {
		return "", errors.New("crc64")
	}
	return fmt.Sprintf("%v", crc64Hash.Sum64()), nil
}

func (p *Vod) vpcUpload(vpcUploadAddress *business.VpcTosUploadAddress, vodUploadFuncRequest *model.VodUploadFuncRequest) error {
	if vpcUploadAddress == nil {
		return errors.New("nil VpcUploadAddress")
	}

	if vpcUploadAddress.GetQuickCompleteMode() == "enable" {
		return nil
	}
	param := model.UploadPartCommon{
		Client:    &http.Client{Timeout: 900 * time.Second},
		FileSize:  vodUploadFuncRequest.Size,
		SpaceName: vodUploadFuncRequest.SpaceName,
	}

	if vpcUploadAddress.GetUploadMode() == "direct" {
		return p.vpcPut(vpcUploadAddress, vodUploadFuncRequest.FilePath, param)
	} else if vpcUploadAddress.GetUploadMode() == "part" {
		return p.vpcPartUpload(vpcUploadAddress.GetPartUploadInfo(), vodUploadFuncRequest.FilePath, param)
	}

	return nil
}

func (p *Vod) vpcPartUpload(partUploadInfo *business.PartUploadInfo, filePath string, param model.UploadPartCommon) error {
	if partUploadInfo == nil {
		return errors.New("empty partInfo")
	}
	size := param.FileSize
	chunkSize := partUploadInfo.GetPartSize()
	totalNum := size / chunkSize
	lastPartSize := size % chunkSize

	if (int64(len(partUploadInfo.GetPartPutUrls())) != totalNum+1 && lastPartSize > 0) ||
		(int64(len(partUploadInfo.GetPartPutUrls())) != totalNum && lastPartSize == 0) {
		return errors.New("mismatch part upload")
	}
	f, err := os.Open(filePath)
	if err != nil {
		return errors.New("file open")
	}
	defer f.Close()

	offset := int64(0)
	partsInfo := &model.VpcUploadPartsInfo{
		Parts: make([]*model.VpcUploadPartInfo, 0),
	}
	for i := 0; i < len(partUploadInfo.GetPartPutUrls())-1; i++ {
		partUrl := partUploadInfo.GetPartPutUrls()[i]
		etag, err := p.vpcPartPut(f, partUrl, offset, chunkSize, param)
		if err != nil {
			return err
		}
		partsInfo.Parts = append(partsInfo.Parts, &model.VpcUploadPartInfo{
			PartNumber: i + 1,
			ETag:       etag,
		})
		offset += chunkSize
	}

	lastChunkSize := size - offset
	etag, err := p.vpcPartPut(f, partUploadInfo.GetPartPutUrls()[totalNum], offset, lastChunkSize, param)
	if err != nil {
		return err
	}
	partsInfo.Parts = append(partsInfo.Parts, &model.VpcUploadPartInfo{
		PartNumber: int(totalNum + 1),
		ETag:       etag,
	})

	return p.vpcPost(partUploadInfo, partsInfo, param)
}

func (p *Vod) vpcPartPut(f *os.File, putUrl string, offset, size int64, param model.UploadPartCommon) (string, error) {
	client := param.Client
	sectionReader := io.NewSectionReader(f, offset, size)
	hashCrc64, err := crc64ecma(sectionReader)
	if err != nil {
		return "", errors.New("crc64")
	}
	_, err = sectionReader.Seek(0, 0)
	if err != nil {
		return "", errors.New("reader seek")
	}

	u, err := url.Parse(putUrl)
	if err != nil {
		return "", errors.New("putUrl is invalid")
	}

	req, err := http.NewRequest("PUT", putUrl, sectionReader)
	if err != nil {
		return "", err
	}

	now := time.Now()
	rsp, err := client.Do(req)
	if err != nil {
		reporter.report(p, p.buildDefaultUploadReport(param.SpaceName, time.Since(now).Microseconds(), 500, 0, "", actionVpcChunkUpload, u.Host, err.Error()))
		return "", err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		logId := rsp.Header.Get("x-tos-request-id")
		var errMsg string
		bts, err := ioutil.ReadAll(rsp.Body)
		if err == nil {
			errMsg = string(bts)
		}
		reporter.report(p, p.buildDefaultUploadReport(param.SpaceName, time.Since(now).Microseconds(), rsp.StatusCode, 0, logId, actionVpcChunkUpload, u.Host, errMsg))
		return "", errors.New("put error:" + logId)
	}

	resCrc64 := rsp.Header.Get("x-tos-hash-crc64ecma")
	if resCrc64 != hashCrc64 {
		return "", errors.New("integrity check failed")
	}
	etag := rsp.Header.Get("ETag")

	return etag, nil
}

func (p *Vod) vpcPut(vpcUploadAddress *business.VpcTosUploadAddress, filePath string, param model.UploadPartCommon) error {
	client := param.Client
	putUrl := vpcUploadAddress.GetPutUrl()
	u, err := url.Parse(putUrl)
	if err != nil {
		return errors.New("putUrl parse failed")
	}
	f, err := os.Open(filePath)
	if err != nil {
		return errors.New("file open")
	}
	defer f.Close()
	hashCrc64, err := crc64ecma(f)
	if err != nil {
		return errors.New("crc64")
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		return errors.New("file seek")
	}

	req, err := http.NewRequest("PUT", putUrl, f)
	if err != nil {
		return err
	}
	for key, value := range vpcUploadAddress.GetPutUrlHeaders() {
		req.Header.Set(key, value)
	}

	now := time.Now()
	rsp, err := client.Do(req)
	if err != nil {
		reporter.report(p, p.buildDefaultUploadReport(param.SpaceName, time.Since(now).Microseconds(), 500, 0, "", actionVpcDirectUpload, u.Host, err.Error()))
		return err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		logId := rsp.Header.Get("x-tos-request-id")
		var errMsg string
		bts, err := ioutil.ReadAll(rsp.Body)
		if err == nil {
			errMsg = string(bts)
		}
		reporter.report(p, p.buildDefaultUploadReport(param.SpaceName, time.Since(now).Microseconds(), rsp.StatusCode, 0, logId, actionVpcDirectUpload, u.Host, errMsg))
		return errors.New("put error:" + logId)
	}

	resCrc64 := rsp.Header.Get("x-tos-hash-crc64ecma")
	if resCrc64 != hashCrc64 {
		return errors.New("integrity check failed")
	}

	return nil
}

func (p *Vod) vpcPost(partUploadInfo *business.PartUploadInfo, partsInfo *model.VpcUploadPartsInfo, param model.UploadPartCommon) error {
	client := param.Client
	postUrl := partUploadInfo.GetCompletePartUrl()
	u, err := url.Parse(postUrl)
	if err != nil {
		return errors.New("postUrl is invalid")
	}

	bodyBytes, err := json.Marshal(partsInfo)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", postUrl, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	for key, value := range partUploadInfo.GetCompleteUrlHeaders() {
		req.Header.Set(key, value)
	}

	now := time.Now()
	rsp, err := client.Do(req)
	if err != nil {
		reporter.report(p, p.buildDefaultUploadReport(param.SpaceName, time.Since(now).Microseconds(), 500, 0, "", actionVpcMergeChunk, u.Host, err.Error()))
		return err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		logId := rsp.Header.Get("x-tos-request-id")
		var errMsg string
		bts, err := ioutil.ReadAll(rsp.Body)
		if err == nil {
			errMsg = string(bts)
		}
		reporter.report(p, p.buildDefaultUploadReport(param.SpaceName, time.Since(now).Microseconds(), rsp.StatusCode, 0, logId, actionVpcMergeChunk, u.Host, errMsg))
		return errors.New("post error:" + logId)
	}

	return nil
}

type UploadPartInfo struct {
	Number int
	OffSet int64
	Size   int64
}

func calculateUploadParts(size, chunkSize int64) (int, []*UploadPartInfo, error) {
	totalNum := size / chunkSize
	if totalNum > 10000 {
		return 0, nil, errors.New("parts over 10000")
	}

	uploadPartInfos := make([]*UploadPartInfo, 0)
	for i := 0; i < int(totalNum); i++ {
		partInfo := &UploadPartInfo{
			Number: i + 1,
			OffSet: int64(i) * chunkSize,
			Size:   chunkSize,
		}
		uploadPartInfos = append(uploadPartInfos, partInfo)
	}

	last := size % chunkSize
	if last != 0 {
		uploadPartInfos[totalNum-1].Size += last
	}
	return int(totalNum), uploadPartInfos, nil
}

type Jobs struct {
	filePath         string
	uploadPartInfo   *UploadPartInfo
	uploadPartCommon *model.UploadPartCommon
	uploadId         string
	client           *http.Client
	storageClass     int32
}

func worker(p *Vod, jobs <-chan *Jobs, results chan<- *model.UploadPartResponse, errChan chan<- *model.UploadPartResponse, quit *int, objectContentType *string, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		if *quit != 0 {
			continue
		}
		fd, err := os.Open(job.filePath)
		if err != nil {
			res := &model.UploadPartResponse{
				UploadPartCommonResponse: model.UploadPartCommonResponse{
					Success: -1,
					Error: model.UploadPartError{
						Code:    -1,
						Error:   err.Error(),
						Message: "open file failed",
					},
				},
				PartNumber: job.uploadPartInfo.Number,
			}
			errChan <- res
			*quit = -1
			return
		}

		data := make([]byte, job.uploadPartInfo.Size)

		_, err = fd.ReadAt(data, job.uploadPartInfo.OffSet)
		if err != nil {
			res := &model.UploadPartResponse{
				UploadPartCommonResponse: model.UploadPartCommonResponse{
					Success: -1,
					Error: model.UploadPartError{
						Code:    -1,
						Error:   err.Error(),
						Message: "read data error",
					},
				},
				PartNumber: job.uploadPartInfo.Number,
			}
			errChan <- res
			*quit = -1
			return
		}

		var (
			resp       *model.UploadPartResponse
			logid      string
			statusCode int
		)
		retry.Do(func() error {
			resp, logid, statusCode, err = p.uploadPart(*job.uploadPartCommon, job.uploadId, job.uploadPartInfo.Number, data, job.client, job.storageClass)
			return err
		}, retry.Attempts(3))
		if err != nil {
			res := &model.UploadPartResponse{
				UploadPartCommonResponse: model.UploadPartCommonResponse{
					Success: -1,
					Error: model.UploadPartError{
						Code:       -1,
						Error:      err.Error(),
						Message:    "upload part fail",
						StatusCode: statusCode,
						Logid:      logid,
					},
				},
				PartNumber: job.uploadPartInfo.Number,
			}
			if e, ok := err.(UploadError); ok {
				res.Error.Code = e.Code
				res.Error.ErrorCode = e.ErrorCode
				res.Error.Message = e.Message
			}
			errChan <- res
			*quit = -1
			return
		}
		results <- resp
		if job.uploadPartInfo.Number == 1 {
			*objectContentType = resp.PayLoad.Meta.ObjectContentType
		}
	}
}

func (p *Vod) chunkUpload(filePath string, param model.UploadPartCommon) error {
	client, size, parallelNum, storageClass := param.Client, param.FileSize, param.ParallelNum, param.StorageClass
	// 1. 计算分片
	totalNum, uploadPartInfos, err := calculateUploadParts(size, param.ChunkSize)
	if err != nil {
		return err
	}

	// 2. Init 初始化分片
	uploadID, err := p.initUploadPartV2(param)
	if err != nil {
		return err
	}

	chJobs := make(chan *Jobs, totalNum)
	chUploadPartRes := make(chan *model.UploadPartResponse, totalNum)
	errChan := make(chan *model.UploadPartResponse, totalNum)
	quitSig := 0
	wg := sync.WaitGroup{}
	wg.Add(parallelNum)
	objectContentType := ""

	now := time.Now()
	// 3. StartWorker
	for w := 1; w <= parallelNum; w++ {
		go worker(p, chJobs, chUploadPartRes, errChan, &quitSig, &objectContentType, &wg)
	}

	// 4. PushJobs
	for i := 0; i < totalNum; i++ {
		job := &Jobs{
			filePath:         filePath,
			uploadPartInfo:   uploadPartInfos[i],
			uploadPartCommon: &param,
			uploadId:         uploadID,
			client:           client,
			storageClass:     storageClass,
		}
		chJobs <- job
	}
	close(chJobs)

	// 5. recive results
	wg.Wait()
	select {
	case v := <-errChan:
		close(chUploadPartRes)
		close(errChan)
		if v.Error.Code == -1 {
			fmt.Printf("Error=%s, Code=%v, Message=%s\n", v.Error.Error, v.Error.Code, v.Error.Message)
			reporter.report(p, p.buildDefaultUploadReport(param.SpaceName, time.Since(now).Microseconds(), v.Error.StatusCode, param.RetryTimes, v.Error.Logid, actionChunkUpload, param.TosHost, err.Error()))
			return fmt.Errorf("Error=%s,Message is %s", v.Error.Error, v.Error.Message)
		}
		return UploadError{
			Code:      v.Error.Code,
			ErrorCode: v.Error.ErrorCode,
			Message:   v.Error.Message,
		}
	default:
		// all parts upload success
		close(errChan)
	}
	uploadPartResponseList := make([]*model.UploadPartResponse, 0)
	for i := 0; i < totalNum; i++ {
		uploadPartResponseList = append(uploadPartResponseList, <-chUploadPartRes)
	}
	close(chUploadPartRes)
	param.ObjectContentType = objectContentType
	return p.uploadMergePartV2(param, uploadID, uploadPartResponseList)
}

func (p *Vod) UploadMergePart(uploadPart model.UploadPartCommon, uploadID string, uploadPartResponseList []*model.UploadPartResponse, client *http.Client, storageClass int32) error {
	url := fmt.Sprintf("https://%s/%s?uploadID=%s", uploadPart.TosHost, uploadPart.Oid, uploadID)
	body, err := p.genMergeBody(uploadPartResponseList)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("PUT", url, bytes.NewReader([]byte(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", uploadPart.Auth)
	req.Header.Set("X-Storage-Mode", "gateway")

	if storageClass == int32(business.StorageClassType_Archive) || storageClass == int32(business.StorageClassType_IA) {
		if storageClass == int32(business.StorageClassType_Archive) {
			req.Header.Set("X-Upload-Storage-Class", "archive")
		}
		if storageClass == int32(business.StorageClassType_IA) {
			req.Header.Set("X-Upload-Storage-Class", "ia")
		}
		if uploadPart.ObjectContentType != "" {
			q := req.URL.Query()
			q.Add("ObjectContentType", uploadPart.ObjectContentType)
			req.URL.RawQuery = q.Encode()
		}
	}

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}
	b, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return err
	}
	res := &model.UploadMergeResponse{}
	if err := json.Unmarshal(b, res); err != nil {
		return err
	}
	if res.Success != 0 {
		return UploadError{
			Code:      res.Error.Code,
			ErrorCode: res.Error.ErrorCode,
			Message:   res.Error.Message,
		}
	}
	return nil
}

func (p *Vod) uploadMergePartV2(param model.UploadPartCommon, uploadID string, uploadPartResponseList []*model.UploadPartResponse) error {
	client, tosHost, storageClass := param.Client, param.TosHost, param.StorageClass
	url := fmt.Sprintf("https://%s/%s?uploadID=%s", tosHost, param.Oid, uploadID)
	body, err := p.genMergeBody(uploadPartResponseList)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("PUT", url, bytes.NewReader([]byte(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", param.Auth)
	req.Header.Set("X-Storage-Mode", "gateway")

	if storageClass == int32(business.StorageClassType_Archive) || storageClass == int32(business.StorageClassType_IA) {
		if storageClass == int32(business.StorageClassType_Archive) {
			req.Header.Set("X-Upload-Storage-Class", "archive")
		}
		if storageClass == int32(business.StorageClassType_IA) {
			req.Header.Set("X-Upload-Storage-Class", "ia")
		}
		if param.ObjectContentType != "" {
			q := req.URL.Query()
			q.Add("ObjectContentType", param.ObjectContentType)
			req.URL.RawQuery = q.Encode()
		}
	}

	now := time.Now()
	rsp, err := client.Do(req)
	if err != nil {
		reporter.report(p, p.buildDefaultUploadReport(param.SpaceName, time.Since(now).Microseconds(), 500, param.RetryTimes, "", actionMergeChunk, tosHost, err.Error()))
		return err
	}
	b, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		reporter.report(p, p.buildDefaultUploadReport(param.SpaceName, time.Since(now).Microseconds(), rsp.StatusCode, param.RetryTimes, rsp.Header.Get(logHeader), actionMergeChunk, tosHost, err.Error()))
		return err
	}
	res := &model.UploadMergeResponse{}
	if err := json.Unmarshal(b, res); err != nil {
		errStr := fmt.Sprintf("unmarshal merge part response failed: %v, got result: %s", err, string(b))
		reporter.report(p, p.buildDefaultUploadReport(param.SpaceName, time.Since(now).Microseconds(), rsp.StatusCode, param.RetryTimes, rsp.Header.Get(logHeader), actionMergeChunk, tosHost, errStr))
		return fmt.Errorf("unmarshal merge upload part response failed: %v, got result: %s", err, string(b))
	}
	if res.Success != 0 {
		return UploadError{
			Code:      res.Error.Code,
			ErrorCode: res.Error.ErrorCode,
			Message:   res.Error.Message,
		}
	}
	return nil
}

func (p *Vod) genMergeBody(uploadPartResponseList []*model.UploadPartResponse) (string, error) {
	if len(uploadPartResponseList) == 0 {
		return "", errors.New("body crc32 empty")
	}
	s := make([]string, len(uploadPartResponseList))
	for _, v := range uploadPartResponseList {
		s[v.PartNumber-1] = fmt.Sprintf("%d:%s", v.PartNumber, v.CheckSum)
	}
	return strings.Join(s, ","), nil
}

func (p *Vod) uploadPart(uploadPart model.UploadPartCommon, uploadID string, partNumber int, data []byte, client *http.Client, storageClass int32) (*model.UploadPartResponse, string, int, error) {
	url := fmt.Sprintf("https://%s/%s?partNumber=%d&uploadID=%s", uploadPart.TosHost, uploadPart.Oid, partNumber, uploadID)
	checkSum := fmt.Sprintf("%08x", crc32.ChecksumIEEE(data))
	req, err := http.NewRequest("PUT", url, bytes.NewReader(data))
	if err != nil {
		return nil, "", -1, err
	}
	req.Header.Set("Content-CRC32", checkSum)
	req.Header.Set("Authorization", uploadPart.Auth)
	req.Header.Set("X-Storage-Mode", "gateway")

	if storageClass == int32(business.StorageClassType_Archive) {
		req.Header.Set("X-Upload-Storage-Class", "archive")
	}
	if storageClass == int32(business.StorageClassType_IA) {
		req.Header.Set("X-Upload-Storage-Class", "ia")
	}

	rsp, err := client.Do(req)
	if err != nil {
		return nil, "", 500, err
	}
	b, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, rsp.Header.Get(logHeader), rsp.StatusCode, err
	}
	res := &model.UploadPartResponse{}
	if err := json.Unmarshal(b, res); err != nil {
		return nil, rsp.Header.Get(logHeader), rsp.StatusCode, fmt.Errorf("unmarshal direct upload response failed: %v, got result: %s", err, string(b))
	}
	if res.Success != 0 {
		return nil, rsp.Header.Get(logHeader), rsp.StatusCode, UploadError{
			Code:      res.Error.Code,
			ErrorCode: res.Error.ErrorCode,
			Message:   res.Error.Message,
		}
	}
	res.PartNumber = partNumber
	res.CheckSum = checkSum
	//return checkSum, res.PayLoad.Meta.ObjectContentType, nil
	return res, rsp.Header.Get(logHeader), rsp.StatusCode, nil
}

func (p *Vod) uploadPartStream(uploadPart model.UploadPartCommon, uploadID string, partNumber int, content io.Reader, client *http.Client, storageClass int32) (*model.UploadPartResponse, string, int, error) {
	url := fmt.Sprintf("https://%s/%s?partNumber=%d&uploadID=%s", uploadPart.TosHost, uploadPart.Oid, partNumber, uploadID)
	checkSum := "Ignore"
	req, err := http.NewRequest("PUT", url, content)
	if err != nil {
		return nil, "", -1, err
	}
	req.Header.Set("Content-CRC32", checkSum)
	req.Header.Set("Authorization", uploadPart.Auth)
	req.Header.Set("X-Storage-Mode", "gateway")

	if storageClass == int32(business.StorageClassType_Archive) {
		req.Header.Set("X-Upload-Storage-Class", "archive")
	}
	if storageClass == int32(business.StorageClassType_IA) {
		req.Header.Set("X-Upload-Storage-Class", "ia")
	}

	rsp, err := client.Do(req)
	if err != nil {
		return nil, "", 500, err
	}
	b, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, rsp.Header.Get(logHeader), rsp.StatusCode, err
	}
	res := &model.UploadPartResponse{}
	if err := json.Unmarshal(b, res); err != nil {
		return nil, rsp.Header.Get(logHeader), rsp.StatusCode, fmt.Errorf("unmarshal init upload part response failed: %v, got result: %s", err, string(b))
	}
	if res.Success != 0 {
		return nil, rsp.Header.Get(logHeader), rsp.StatusCode, UploadError{
			Code:      res.Error.Code,
			ErrorCode: res.Error.ErrorCode,
			Message:   res.Error.Message,
		}
	}
	res.PartNumber = partNumber
	res.CheckSum = checkSum
	//return checkSum, res.PayLoad.Meta.ObjectContentType, nil
	return res, rsp.Header.Get(logHeader), rsp.StatusCode, nil
}

func (p *Vod) BuildVodCommonUploadInfo(resp *response.VodApplyUploadInfoResponse) (*model.UploadCommonInfo, error) {
	if resp == nil || resp.ResponseMetadata == nil || resp.Result == nil {
		return nil, errors.New("empty input")
	}

	if resp.ResponseMetadata.Error != nil && resp.ResponseMetadata.Error.Code != "0" {
		return nil, fmt.Errorf("%+v", resp.ResponseMetadata.Error)
	}

	uploadCommonInfo := &model.UploadCommonInfo{
		Client: &http.Client{},
	}

	// using candidate address first
	candidateUploadAddress := resp.GetResult().GetData().GetCandidateUploadAddresses()
	var allUploadAddress []*business.UploadAddress
	if candidateUploadAddress != nil {
		allUploadAddress = append(allUploadAddress, candidateUploadAddress.GetMainUploadAddresses()...)
		allUploadAddress = append(allUploadAddress, candidateUploadAddress.GetBackupUploadAddresses()...)
		allUploadAddress = append(allUploadAddress, candidateUploadAddress.GetFallbackUploadAddresses()...)
	}
	uploadAddress := resp.GetResult().GetData().GetUploadAddress()
	if uploadAddress == nil || len(allUploadAddress) == 0 {
		return nil, errors.New("empty address")
	}

	if uploadAddress != nil {
		if len(uploadAddress.GetUploadHosts()) == 0 {
			return nil, fmt.Errorf("no tos host found")
		}
		if len(uploadAddress.GetStoreInfos()) == 0 || (uploadAddress.GetStoreInfos()[0] == nil) {
			return nil, fmt.Errorf("no store info found")
		}

		uploadCommonInfo.Hosts = uploadAddress.GetUploadHosts()
		uploadCommonInfo.Oid = uploadAddress.StoreInfos[0].GetStoreUri()
		uploadCommonInfo.SessionKey = uploadAddress.GetSessionKey()
		uploadCommonInfo.Auth = uploadAddress.GetStoreInfos()[0].GetAuth()
	}

	if len(allUploadAddress) > 0 {
		candidateHosts := make([]string, 0)
		for _, candidateAddress := range allUploadAddress {
			candidateHosts = append(candidateHosts, candidateAddress.UploadHosts...)
		}
		if len(candidateHosts) > 0 {
			uploadCommonInfo.Hosts = candidateHosts
		}
	}

	return uploadCommonInfo, nil
}

func (p *Vod) checkUploadCommonInfo(uploadInfo *model.UploadCommonInfo) error {
	if uploadInfo == nil || uploadInfo.Auth == "" || uploadInfo.Oid == "" ||
		len(uploadInfo.Hosts) == 0 || uploadInfo.Client == nil {
		return UploadError{
			Message: "wrong upload common info",
		}
	}
	return nil
}

func (p *Vod) pickUploadHost(uploadInfo *model.UploadCommonInfo) string {
	index := 0
	if uploadInfo.PreferredHostIndex > 0 &&
		uploadInfo.PreferredHostIndex < len(uploadInfo.Hosts) {
		index = uploadInfo.PreferredHostIndex
	}
	host := uploadInfo.Hosts[index]
	return host
}

func (p *Vod) CreateMultipartUpload(input *model.CreateMultipartUploadInput) (string, error) {
	if input == nil {
		return "", UploadError{
			Message: "empty input",
		}
	}
	err := p.checkUploadCommonInfo(input.UploadCommonInfo)
	if err != nil {
		return "", err
	}

	host := p.pickUploadHost(input.UploadCommonInfo)
	param := model.UploadPartCommon{
		SpaceName:    input.UploadCommonInfo.SpaceName,
		Client:       input.UploadCommonInfo.Client,
		TosHost:      host,
		Oid:          input.UploadCommonInfo.Oid,
		Auth:         input.UploadCommonInfo.Auth,
		StorageClass: input.UploadCommonInfo.StorageClass,
	}
	return p.initUploadPartV2(param)
}

func (p *Vod) UploadPart(input *model.UploadPartInput) (*model.UploadPartResponse, error) {
	if input == nil {
		return nil, UploadError{
			Message: "empty input",
		}
	}
	err := p.checkUploadCommonInfo(input.UploadCommonInfo)
	if err != nil {
		return nil, err
	}

	host := p.pickUploadHost(input.UploadCommonInfo)
	partInfo := model.UploadPartCommon{
		TosHost: host,
		Oid:     input.UploadCommonInfo.Oid,
		Auth:    input.UploadCommonInfo.Auth,
	}

	var (
		resp   *model.UploadPartResponse
		logId  string
		status int
	)

	now := time.Now()
	if input.Data != nil && len(input.Data) != 0 {
		resp, logId, status, err = p.uploadPart(partInfo, input.UploadId, int(input.PartNumber), input.Data,
			input.UploadCommonInfo.Client, input.UploadCommonInfo.StorageClass)
	} else if input.Content != nil {
		resp, logId, status, err = p.uploadPartStream(partInfo, input.UploadId, int(input.PartNumber), input.Content,
			input.UploadCommonInfo.Client, input.UploadCommonInfo.StorageClass)
	} else {
		return nil, errors.New("nil data&content")
	}
	if err != nil {
		reporter.report(p, p.buildDefaultUploadReport(input.UploadCommonInfo.SpaceName, time.Since(now).Microseconds(), status, 0, logId, actionChunkUpload, host, err.Error()))
	}
	return resp, err
}

func (p *Vod) CompleteMultipartUpload(input *model.CompleteMultipartUploadInput) error {
	if input == nil {
		return UploadError{
			Message: "empty input",
		}
	}
	err := p.checkUploadCommonInfo(input.UploadCommonInfo)
	if err != nil {
		return err
	}

	host := p.pickUploadHost(input.UploadCommonInfo)
	uploadPart := model.UploadPartCommon{
		Client:       input.UploadCommonInfo.Client,
		StorageClass: input.UploadCommonInfo.StorageClass,
		TosHost:      host,
		Oid:          input.UploadCommonInfo.Oid,
		Auth:         input.UploadCommonInfo.Auth,
	}
	return p.uploadMergePartV2(uploadPart, input.UploadId, input.PartList)
}

func (p *Vod) PutObject(input *model.PutObjectInput) error {
	if input == nil {
		return UploadError{
			Message: "empty input",
		}
	}
	err := p.checkUploadCommonInfo(input.UploadCommonInfo)
	if err != nil {
		return err
	}

	host := p.pickUploadHost(input.UploadCommonInfo)
	param := model.UploadPartCommon{
		Client:       input.UploadCommonInfo.Client,
		TosHost:      host,
		Oid:          input.UploadCommonInfo.Oid,
		Auth:         input.UploadCommonInfo.Auth,
		StorageClass: input.UploadCommonInfo.StorageClass,
		SpaceName:    input.UploadCommonInfo.SpaceName,
	}
	if input.Data != nil && len(input.Data) != 0 {
		return p.directUpload(input.Data, param)
	} else if input.Content != nil {
		return p.directUploadStream(input.Content, param)
	}
	return errors.New("nil data and content")
}

func (p *Vod) InitUploadPart(tosHost string, oid string, auth string, client *http.Client, storageClass int32) (string, error) {
	url := fmt.Sprintf("https://%s/%s?uploads", tosHost, oid)
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", auth)
	req.Header.Set("X-Storage-Mode", "gateway")

	if storageClass == int32(business.StorageClassType_Archive) {
		req.Header.Set("X-Upload-Storage-Class", "archive")
	}
	if storageClass == int32(business.StorageClassType_IA) {
		req.Header.Set("X-Upload-Storage-Class", "ia")
	}

	rsp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	b, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return "", err
	}
	res := &model.InitPartResponse{}
	if err := json.Unmarshal(b, res); err != nil {
		return "", err
	}
	if res.Success != 0 {
		return "", UploadError{
			Code:      res.Error.Code,
			ErrorCode: res.Error.ErrorCode,
			Message:   res.Error.Message,
		}
	}
	return res.PayLoad.UploadID, nil
}

func (p *Vod) initUploadPartV2(param model.UploadPartCommon) (string, error) {
	tosHost, oid, auth, client, storageClass := param.TosHost, param.Oid, param.Auth, param.Client, param.StorageClass
	url := fmt.Sprintf("https://%s/%s?uploads", tosHost, oid)
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", auth)
	req.Header.Set("X-Storage-Mode", "gateway")

	if storageClass == int32(business.StorageClassType_Archive) {
		req.Header.Set("X-Upload-Storage-Class", "archive")
	}
	if storageClass == int32(business.StorageClassType_IA) {
		req.Header.Set("X-Upload-Storage-Class", "ia")
	}

	now := time.Now()
	rsp, err := client.Do(req)
	if err != nil {
		reporter.report(p, p.buildDefaultUploadReport(param.SpaceName, time.Since(now).Microseconds(), 500, param.RetryTimes, "", actionInitChunk, tosHost, err.Error()))
		return "", err
	}
	b, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		reporter.report(p, p.buildDefaultUploadReport(param.SpaceName, time.Since(now).Microseconds(), rsp.StatusCode, param.RetryTimes, rsp.Header.Get(logHeader), actionInitChunk, tosHost, err.Error()))
		return "", err
	}
	res := &model.InitPartResponse{}
	if err := json.Unmarshal(b, res); err != nil {
		err = fmt.Errorf("unmarshal init upload part response failed: %v, got result: %s", err, string(b))
		reporter.report(p, p.buildDefaultUploadReport(param.SpaceName, time.Since(now).Microseconds(), rsp.StatusCode, param.RetryTimes, rsp.Header.Get(logHeader), actionInitChunk, tosHost, err.Error()))
		return "", err
	}
	if res.Success != 0 {
		return "", UploadError{
			Code:      res.Error.Code,
			ErrorCode: res.Error.ErrorCode,
			Message:   res.Error.Message,
		}
	}
	return res.PayLoad.UploadID, nil
}

func (p *Vod) MoveObjectCrossSpace(req *request.VodSubmitMoveObjectTaskRequest, cycleNum int) (*response.VodQueryMoveObjectTaskInfoResponse, int, error) {
	submitResp, submitStatus, err := p.SubmitMoveObjectTask(req)
	if err != nil || submitStatus != http.StatusOK {
		return &response.VodQueryMoveObjectTaskInfoResponse{ResponseMetadata: submitResp.GetResponseMetadata()}, submitStatus, err
	}

	if cycleNum == 0 {
		cycleNum = 600
	}
	for i := 0; i < cycleNum; i++ {
		queryResp, queryStatus, err := p.QueryMoveObjectTaskInfo(&request.VodQueryMoveObjectTaskInfoRequest{
			TaskId:      submitResp.GetResult().GetData().GetTaskId(),
			SourceSpace: req.SourceSpace,
			TargetSpace: req.TargetSpace,
		})
		if err != nil || queryStatus != http.StatusOK {
			return queryResp, queryStatus, err
		}
		if queryResp.GetResult().GetData().GetState() == "success" || queryResp.GetResult().GetData().GetState() == "failed" {
			return queryResp, queryStatus, err
		}
		time.Sleep(time.Second)
	}
	return nil, http.StatusGatewayTimeout, errors.New(fmt.Sprintf("task run time out, requestId is %s , please retry or contact volc assistant", submitResp.GetResponseMetadata().GetRequestId()))
}

func (p *Vod) UploadMediaStreamWithCallback(mediaRequset *model.VodStreamUploadRequest) (*response.VodCommitUploadInfoResponse, int, error) {
	mediaRequset.FileType = "media"
	return p.StreamUploadInner(mediaRequset)
}

func (p *Vod) UploadMaterialStreamWithCallback(mediaRequset *model.VodStreamUploadRequest) (*response.VodCommitUploadInfoResponse, int, error) {
	return p.StreamUploadInner(mediaRequset)
}

func (p *Vod) UploadObjectStreamWithCallback(mediaRequset *model.VodStreamUploadRequest) (*response.VodCommitUploadInfoResponse, int, error) {
	mediaRequset.FileType = "object"
	return p.StreamUploadInner(mediaRequset)
}

func (p *Vod) StreamUploadInner(streamUploadInnerRequest *model.VodStreamUploadRequest) (r *response.VodCommitUploadInfoResponse, s int, e error) {
	now := time.Now()
	defer func() {
		var (
			msg  string
			code = 200
		)
		if e != nil {
			msg = e.Error()
			code = 500
		}
		reporter.report(p, p.buildDefaultUploadReport(streamUploadInnerRequest.SpaceName, time.Since(now).Microseconds(), code, 0, "", actionFinishUpload, "", msg))
	}()
	req := &model.VodStreamUploadRequest{
		Content:           streamUploadInnerRequest.Content,
		Size:              streamUploadInnerRequest.Size,
		SpaceName:         streamUploadInnerRequest.SpaceName,
		FileType:          streamUploadInnerRequest.FileType,
		FileName:          streamUploadInnerRequest.FileName,
		FileExtension:     streamUploadInnerRequest.FileExtension,
		StorageClass:      streamUploadInnerRequest.StorageClass,
		ClientNetWorkMode: streamUploadInnerRequest.ClientNetWorkMode,
		ClientIDCMode:     streamUploadInnerRequest.ClientIDCMode,
		UploadHostPrefer:  streamUploadInnerRequest.UploadHostPrefer,
		ChunkSize:         streamUploadInnerRequest.ChunkSize,
	}
	logId, sessionKey, err, code := p.StreamUpload(req)
	if err != nil {
		return p.fillCommitUploadInfoResponseWhenError(logId, err.Error()), code, err
	}

	commitRequest := &request.VodCommitUploadInfoRequest{
		SpaceName:       streamUploadInnerRequest.SpaceName,
		SessionKey:      sessionKey,
		CallbackArgs:    streamUploadInnerRequest.CallbackArgs,
		Functions:       streamUploadInnerRequest.Functions,
		VodUploadSource: streamUploadInnerRequest.VodUploadSource,
		ExpireTime:      streamUploadInnerRequest.ExpireTime,
	}

	now1 := time.Now()
	commitResp, code, err := p.CommitUploadInfo(commitRequest)
	if err != nil {
		if commitResp == nil {
			reporter.report(p, p.buildDefaultUploadReport(streamUploadInnerRequest.SpaceName, time.Since(now1).Microseconds(), code, 0, "", actionCommitUploadInfo, "", err.Error()))
		}
		return commitResp, code, err
	}
	return commitResp, code, nil
}

func (p *Vod) StreamUpload(vodStreamUploadRequest *model.VodStreamUploadRequest) (string, string, error, int) {
	if vodStreamUploadRequest.ChunkSize == 0 {
		vodStreamUploadRequest.ChunkSize = consts.MinChunckSize
	}
	if vodStreamUploadRequest.ChunkSize < consts.StreamMinChunkSize {
		return "", "", errors.New("chunk size must be greater than 5MB"), 0
	}
	if vodStreamUploadRequest.Content == nil {
		return "", "", errors.New("content is nil"), 0
	}

	applyRequest := &request.VodApplyUploadInfoRequest{
		SpaceName:         vodStreamUploadRequest.SpaceName,
		FileType:          vodStreamUploadRequest.FileType,
		FileName:          vodStreamUploadRequest.FileName,
		FileExtension:     vodStreamUploadRequest.FileExtension,
		StorageClass:      vodStreamUploadRequest.StorageClass,
		ClientNetWorkMode: vodStreamUploadRequest.ClientNetWorkMode,
		ClientIDCMode:     vodStreamUploadRequest.ClientIDCMode,
		NeedFallback:      true, // default set
		UploadHostPrefer:  vodStreamUploadRequest.UploadHostPrefer,
		FileSize:          float64(vodStreamUploadRequest.Size),
	}

	now := time.Now()
	resp, code, err := p.ApplyUploadInfo(applyRequest)
	logId := resp.GetResponseMetadata().GetRequestId()
	if err != nil {
		if resp == nil {
			reporter.report(p, p.buildDefaultUploadReport(vodStreamUploadRequest.SpaceName, time.Since(now).Microseconds(), code, 0, logId, actionApplyUploadInfo, "", err.Error()))
		}
		return logId, "", err, code
	}

	if resp.ResponseMetadata.Error != nil && resp.ResponseMetadata.Error.Code != "0" {
		return logId, "", fmt.Errorf("%+v", resp.ResponseMetadata.Error), code
	}

	//vpc upload
	if vpcUploadAddress := resp.GetResult().GetData().GetVpcTosUploadAddress(); vpcUploadAddress != nil {
		err := p.vpcUploadStream(vpcUploadAddress, vodStreamUploadRequest)
		if err != nil {
			return logId, "", err, http.StatusBadRequest
		}
		uploadAddress := resp.GetResult().GetData().GetUploadAddress()
		oid := uploadAddress.StoreInfos[0].GetStoreUri()
		sessionKey := uploadAddress.GetSessionKey()
		return oid, sessionKey, nil, http.StatusOK
	}

	uploadCommonInfo, err := p.BuildVodCommonUploadInfo(resp)
	if err != nil {
		return logId, "", fmt.Errorf("build common upload info error:%+v", err), code
	}
	uploadCommonInfo.StorageClass = vodStreamUploadRequest.StorageClass
	uploadCommonInfo.SpaceName = vodStreamUploadRequest.SpaceName

	err = p.StreamUploadContent(&model.UploadContentParam{
		UploadCommonInfo: uploadCommonInfo,
		ChunkSize:        vodStreamUploadRequest.ChunkSize,
		Size:             vodStreamUploadRequest.Size,
		Content:          vodStreamUploadRequest.Content,
	})
	if err != nil {
		return logId, "", err, http.StatusBadRequest
	}

	return uploadCommonInfo.Oid, uploadCommonInfo.SessionKey, nil, http.StatusOK
}

func (p *Vod) StreamUploadContent(param *model.UploadContentParam) error {
	if param.UploadCommonInfo == nil {
		return errors.New("nil upload common info")
	}
	err := p.checkUploadCommonInfo(param.UploadCommonInfo)
	if err != nil {
		return err
	}

	uploadCommonInfo := param.UploadCommonInfo
	if param.Size == 0 {
		return p.StreamUploadContentInChunk(param)
	}

	// part upload
	if param.Size > param.ChunkSize {
		uploadId, err := p.CreateMultipartUpload(&model.CreateMultipartUploadInput{
			UploadCommonInfo: uploadCommonInfo,
		})
		if err != nil {
			return fmt.Errorf("CreateMultipartUpload Error:%v", err)
		}

		probeBytes := make([]byte, 1)
		var index int64 = 1
		partList := make([]*model.UploadPartResponse, 0)
		for {
			_, err := io.ReadFull(param.Content, probeBytes)
			if err == io.EOF {
				break
			}

			chunkReader := io.LimitReader(io.MultiReader(bytes.NewReader(probeBytes), param.Content), param.ChunkSize)
			partInfo, err := p.UploadPart(&model.UploadPartInput{
				UploadCommonInfo: uploadCommonInfo,
				PartNumber:       index,
				UploadId:         uploadId,
				Content:          chunkReader,
			})
			if err != nil {
				return fmt.Errorf("UploadPart Error:%v", err)
			}
			partList = append(partList, partInfo)
			index++
		}

		err = p.CompleteMultipartUpload(&model.CompleteMultipartUploadInput{
			UploadCommonInfo: uploadCommonInfo,
			UploadId:         uploadId,
			PartList:         partList,
		})
		if err != nil {
			return fmt.Errorf("CompleteMultipartUpload Error:%v", err)
		}
	} else {
		return p.PutObject(&model.PutObjectInput{
			UploadCommonInfo: uploadCommonInfo,
			Content:          param.Content,
		})
	}

	return nil
}

func (p *Vod) StreamUploadContentInChunk(param *model.UploadContentParam) error {
	if param.UploadCommonInfo == nil {
		return errors.New("nil upload common info")
	}
	err := p.checkUploadCommonInfo(param.UploadCommonInfo)
	if err != nil {
		return err
	}
	uploadCommonInfo := param.UploadCommonInfo

	firstReader := io.LimitReader(param.Content, param.ChunkSize)
	data, err := ioutil.ReadAll(firstReader)
	if err != nil {
		return err
	}

	if int64(len(data)) < param.ChunkSize {
		return p.PutObject(&model.PutObjectInput{
			UploadCommonInfo: uploadCommonInfo,
			Data:             data,
		})
	} else {
		uploadId, err := p.CreateMultipartUpload(&model.CreateMultipartUploadInput{
			UploadCommonInfo: uploadCommonInfo,
		})
		if err != nil {
			return fmt.Errorf("CreateMultipartUpload Error:%v", err)
		}

		var index int64 = 1
		partList := make([]*model.UploadPartResponse, 0)
		for {
			partInfo, err := p.UploadPart(&model.UploadPartInput{
				UploadCommonInfo: uploadCommonInfo,
				PartNumber:       index,
				UploadId:         uploadId,
				Data:             data,
			})
			if err != nil {
				return fmt.Errorf("UploadPart Error:%v", err)
			}
			partList = append(partList, partInfo)
			index++

			limitReader := io.LimitReader(param.Content, param.ChunkSize)
			data, err = ioutil.ReadAll(limitReader)
			if err != nil {
				return err
			}

			if len(data) == 0 {
				break
			}
		}

		err = p.CompleteMultipartUpload(&model.CompleteMultipartUploadInput{
			UploadCommonInfo: uploadCommonInfo,
			UploadId:         uploadId,
			PartList:         partList,
		})
		if err != nil {
			return fmt.Errorf("CompleteMultipartUpload Error:%v", err)
		}
	}

	return nil
}

func (p *Vod) vpcUploadStream(vpcUploadAddress *business.VpcTosUploadAddress, vodStreamUploadRequest *model.VodStreamUploadRequest) error {
	if vpcUploadAddress == nil {
		return errors.New("nil VpcUploadAddress")
	}

	if vpcUploadAddress.GetQuickCompleteMode() == "enable" {
		return nil
	}
	param := model.UploadPartCommon{
		Client:    &http.Client{Timeout: 900 * time.Second},
		FileSize:  vodStreamUploadRequest.Size,
		SpaceName: vodStreamUploadRequest.SpaceName,
	}

	if vpcUploadAddress.GetUploadMode() == "direct" {
		return p.vpcPutStream(vpcUploadAddress, vodStreamUploadRequest.Content, param)
	} else if vpcUploadAddress.GetUploadMode() == "part" {
		return p.vpcPartUploadStream(vpcUploadAddress.GetPartUploadInfo(), vodStreamUploadRequest.Content, param)
	}

	return nil
}

func (p *Vod) vpcPutStream(vpcUploadAddress *business.VpcTosUploadAddress, content io.Reader, param model.UploadPartCommon) error {
	client := param.Client
	putUrl := vpcUploadAddress.GetPutUrl()
	u, err := url.Parse(putUrl)
	if err != nil {
		return errors.New("putUrl parse failed")
	}
	req, err := http.NewRequest("PUT", putUrl, content)
	if err != nil {
		return err
	}
	for key, value := range vpcUploadAddress.GetPutUrlHeaders() {
		req.Header.Set(key, value)
	}

	now := time.Now()
	rsp, err := client.Do(req)
	if err != nil {
		reporter.report(p, p.buildDefaultUploadReport(param.SpaceName, time.Since(now).Microseconds(), 500, 0, "", actionVpcDirectUpload, u.Host, err.Error()))
		return err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		logId := rsp.Header.Get("x-tos-request-id")
		var errMsg string
		bts, err := ioutil.ReadAll(rsp.Body)
		if err == nil {
			errMsg = string(bts)
		}
		reporter.report(p, p.buildDefaultUploadReport(param.SpaceName, time.Since(now).Microseconds(), rsp.StatusCode, 0, logId, actionVpcDirectUpload, u.Host, errMsg))
		return errors.New("put error:" + logId)
	}

	return nil
}

func (p *Vod) vpcPartPutStream(putUrl string, content io.Reader, param model.UploadPartCommon) (string, error) {
	client := param.Client
	u, err := url.Parse(putUrl)
	if err != nil {
		return "", errors.New("putUrl is invalid")
	}

	req, err := http.NewRequest("PUT", putUrl, content)
	if err != nil {
		return "", err
	}

	now := time.Now()
	rsp, err := client.Do(req)
	if err != nil {
		reporter.report(p, p.buildDefaultUploadReport(param.SpaceName, time.Since(now).Microseconds(), 500, 0, "", actionVpcChunkUpload, u.Host, err.Error()))
		return "", err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		logId := rsp.Header.Get("x-tos-request-id")
		var errMsg string
		bts, err := ioutil.ReadAll(rsp.Body)
		if err == nil {
			errMsg = string(bts)
		}
		reporter.report(p, p.buildDefaultUploadReport(param.SpaceName, time.Since(now).Microseconds(), rsp.StatusCode, 0, logId, actionVpcChunkUpload, u.Host, errMsg))
		return "", errors.New("put error:" + logId)
	}

	etag := rsp.Header.Get("ETag")
	return etag, nil
}

func (p *Vod) vpcPartUploadStream(partUploadInfo *business.PartUploadInfo, content io.Reader, param model.UploadPartCommon) error {
	if partUploadInfo == nil {
		return errors.New("empty partInfo")
	}
	size := param.FileSize
	chunkSize := partUploadInfo.GetPartSize()
	totalNum := size / chunkSize

	if int64(len(partUploadInfo.GetPartPutUrls())) != totalNum+1 {
		return errors.New("mismatch part upload")
	}

	partsInfo := &model.VpcUploadPartsInfo{
		Parts: make([]*model.VpcUploadPartInfo, 0),
	}
	for i := 0; i < len(partUploadInfo.GetPartPutUrls()); i++ {
		partUrl := partUploadInfo.GetPartPutUrls()[i]
		etag, err := p.vpcPartPutStream(partUrl, io.LimitReader(content, chunkSize), param)
		if err != nil {
			return err
		}
		partsInfo.Parts = append(partsInfo.Parts, &model.VpcUploadPartInfo{
			PartNumber: i + 1,
			ETag:       etag,
		})
	}

	probeBytes := make([]byte, 1)
	_, err := io.ReadFull(content, probeBytes)
	if err != io.EOF {
		return errors.New("size & content mismatch")
	}

	return p.vpcPost(partUploadInfo, partsInfo, param)
}

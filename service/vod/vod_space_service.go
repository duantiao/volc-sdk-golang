// Code generated by protoc-gen-volcengine-sdk
// source: VodSpaceService
// DO NOT EDIT!

package vod

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/pkg/errors"

	"github.com/volcengine/volc-sdk-golang/service/vod/models/request"
	"github.com/volcengine/volc-sdk-golang/service/vod/models/response"
)

// DeleteSpace
/*
 * @param *request.VodDeleteSpaceRequest
 * @return *response.VodDeleteSpaceResponse, int, error
 */
func (p *Vod) DeleteSpace(req *request.VodDeleteSpaceRequest) (*response.VodDeleteSpaceResponse, int, error) {
	query := url.Values{}
	form := url.Values{}
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
		return nil, 0, err
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
				return nil, 0, e2
			}
			sv = string(v2)
		}
		query.Set(k, sv)
		form.Add(k, sv)
	}

	respBody, status, err := p.Query("DeleteSpace", query)

	output := &response.VodDeleteSpaceResponse{}
	unmarshaler := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
	errUnmarshal := unmarshaler.Unmarshal(respBody, output)
	if err != nil || status != http.StatusOK {
		// if exist http err,check whether the respBody's type is defined struct,
		// if it is ,
		// return struct,
		// otherwise return nil body
		// if httpCode is not 200,check whether the respBody's type is defined struct,
		// if it is ,
		// use errorCode as err and return struct,
		// otherwise use respBody string as error and return
		if errUnmarshal != nil || len(output.GetResponseMetadata().GetError().GetCode()) == 0 {
			if err == nil {
				err = errors.New(string(respBody))
			}
			return nil, status, err
		} else {
			return output, status, errors.New(output.GetResponseMetadata().GetError().GetCode())
		}
	}
	return output, status, nil
}

// CreateSpace
/*
 * @param *request.VodCreateSpaceRequest
 * @return *response.VodCreateSpaceResponse, int, error
 */
func (p *Vod) CreateSpace(req *request.VodCreateSpaceRequest) (*response.VodCreateSpaceResponse, int, error) {
	query := url.Values{}
	form := url.Values{}
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
		return nil, 0, err
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
				return nil, 0, e2
			}
			sv = string(v2)
		}
		query.Set(k, sv)
		form.Add(k, sv)
	}

	respBody, status, err := p.Query("CreateSpace", query)

	output := &response.VodCreateSpaceResponse{}
	unmarshaler := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
	errUnmarshal := unmarshaler.Unmarshal(respBody, output)
	if err != nil || status != http.StatusOK {
		// if exist http err,check whether the respBody's type is defined struct,
		// if it is ,
		// return struct,
		// otherwise return nil body
		// if httpCode is not 200,check whether the respBody's type is defined struct,
		// if it is ,
		// use errorCode as err and return struct,
		// otherwise use respBody string as error and return
		if errUnmarshal != nil || len(output.GetResponseMetadata().GetError().GetCode()) == 0 {
			if err == nil {
				err = errors.New(string(respBody))
			}
			return nil, status, err
		} else {
			return output, status, errors.New(output.GetResponseMetadata().GetError().GetCode())
		}
	}
	return output, status, nil
}

// ListSpace
/*
 * @param *request.VodListSpaceRequest
 * @return *response.VodListSpaceResponse, int, error
 */
func (p *Vod) ListSpace(req *request.VodListSpaceRequest) (*response.VodListSpaceResponse, int, error) {
	query := url.Values{}
	form := url.Values{}
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
		return nil, 0, err
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
				return nil, 0, e2
			}
			sv = string(v2)
		}
		query.Set(k, sv)
		form.Add(k, sv)
	}

	respBody, status, err := p.Query("ListSpace", query)

	output := &response.VodListSpaceResponse{}
	unmarshaler := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
	errUnmarshal := unmarshaler.Unmarshal(respBody, output)
	if err != nil || status != http.StatusOK {
		// if exist http err,check whether the respBody's type is defined struct,
		// if it is ,
		// return struct,
		// otherwise return nil body
		// if httpCode is not 200,check whether the respBody's type is defined struct,
		// if it is ,
		// use errorCode as err and return struct,
		// otherwise use respBody string as error and return
		if errUnmarshal != nil || len(output.GetResponseMetadata().GetError().GetCode()) == 0 {
			if err == nil {
				err = errors.New(string(respBody))
			}
			return nil, status, err
		} else {
			return output, status, errors.New(output.GetResponseMetadata().GetError().GetCode())
		}
	}
	return output, status, nil
}

// GetSpaceDetail
/*
 * @param *request.VodGetSpaceDetailRequest
 * @return *response.VodGetSpaceDetailResponse, int, error
 */
func (p *Vod) GetSpaceDetail(req *request.VodGetSpaceDetailRequest) (*response.VodGetSpaceDetailResponse, int, error) {
	query := url.Values{}
	form := url.Values{}
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
		return nil, 0, err
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
				return nil, 0, e2
			}
			sv = string(v2)
		}
		query.Set(k, sv)
		form.Add(k, sv)
	}

	respBody, status, err := p.Query("GetSpaceDetail", query)

	output := &response.VodGetSpaceDetailResponse{}
	unmarshaler := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
	errUnmarshal := unmarshaler.Unmarshal(respBody, output)
	if err != nil || status != http.StatusOK {
		// if exist http err,check whether the respBody's type is defined struct,
		// if it is ,
		// return struct,
		// otherwise return nil body
		// if httpCode is not 200,check whether the respBody's type is defined struct,
		// if it is ,
		// use errorCode as err and return struct,
		// otherwise use respBody string as error and return
		if errUnmarshal != nil || len(output.GetResponseMetadata().GetError().GetCode()) == 0 {
			if err == nil {
				err = errors.New(string(respBody))
			}
			return nil, status, err
		} else {
			return output, status, errors.New(output.GetResponseMetadata().GetError().GetCode())
		}
	}
	return output, status, nil
}

// UpdateSpace
/*
 * @param *request.VodUpdateSpaceRequest
 * @return *response.VodUpdateSpaceResponse, int, error
 */
func (p *Vod) UpdateSpace(req *request.VodUpdateSpaceRequest) (*response.VodUpdateSpaceResponse, int, error) {
	query := url.Values{}
	form := url.Values{}
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
		return nil, 0, err
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
				return nil, 0, e2
			}
			sv = string(v2)
		}
		query.Set(k, sv)
		form.Add(k, sv)
	}

	respBody, status, err := p.Query("UpdateSpace", query)

	output := &response.VodUpdateSpaceResponse{}
	unmarshaler := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
	errUnmarshal := unmarshaler.Unmarshal(respBody, output)
	if err != nil || status != http.StatusOK {
		// if exist http err,check whether the respBody's type is defined struct,
		// if it is ,
		// return struct,
		// otherwise return nil body
		// if httpCode is not 200,check whether the respBody's type is defined struct,
		// if it is ,
		// use errorCode as err and return struct,
		// otherwise use respBody string as error and return
		if errUnmarshal != nil || len(output.GetResponseMetadata().GetError().GetCode()) == 0 {
			if err == nil {
				err = errors.New(string(respBody))
			}
			return nil, status, err
		} else {
			return output, status, errors.New(output.GetResponseMetadata().GetError().GetCode())
		}
	}
	return output, status, nil
}

// UpdateSpaceUploadConfig
/*
 * @param *request.VodUpdateSpaceUploadConfigRequest
 * @return *response.VodUpdateSpaceUploadConfigResponse, int, error
 */
func (p *Vod) UpdateSpaceUploadConfig(req *request.VodUpdateSpaceUploadConfigRequest) (*response.VodUpdateSpaceUploadConfigResponse, int, error) {
	query := url.Values{}
	form := url.Values{}
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
		return nil, 0, err
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
				return nil, 0, e2
			}
			sv = string(v2)
		}
		query.Set(k, sv)
		form.Add(k, sv)
	}

	respBody, status, err := p.Query("UpdateSpaceUploadConfig", query)

	output := &response.VodUpdateSpaceUploadConfigResponse{}
	unmarshaler := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
	errUnmarshal := unmarshaler.Unmarshal(respBody, output)
	if err != nil || status != http.StatusOK {
		// if exist http err,check whether the respBody's type is defined struct,
		// if it is ,
		// return struct,
		// otherwise return nil body
		// if httpCode is not 200,check whether the respBody's type is defined struct,
		// if it is ,
		// use errorCode as err and return struct,
		// otherwise use respBody string as error and return
		if errUnmarshal != nil || len(output.GetResponseMetadata().GetError().GetCode()) == 0 {
			if err == nil {
				err = errors.New(string(respBody))
			}
			return nil, status, err
		} else {
			return output, status, errors.New(output.GetResponseMetadata().GetError().GetCode())
		}
	}
	return output, status, nil
}

// DescribeUploadSpaceConfig
/*
 * @param *request.VodDescribeUploadSpaceConfigRequest
 * @return *response.VodDescribeUploadSpaceConfigResponse, int, error
 */
func (p *Vod) DescribeUploadSpaceConfig(req *request.VodDescribeUploadSpaceConfigRequest) (*response.VodDescribeUploadSpaceConfigResponse, int, error) {
	query := url.Values{}
	form := url.Values{}
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
		return nil, 0, err
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
				return nil, 0, e2
			}
			sv = string(v2)
		}
		query.Set(k, sv)
		form.Add(k, sv)
	}

	respBody, status, err := p.Query("DescribeUploadSpaceConfig", query)

	output := &response.VodDescribeUploadSpaceConfigResponse{}
	unmarshaler := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
	errUnmarshal := unmarshaler.Unmarshal(respBody, output)
	if err != nil || status != http.StatusOK {
		// if exist http err,check whether the respBody's type is defined struct,
		// if it is ,
		// return struct,
		// otherwise return nil body
		// if httpCode is not 200,check whether the respBody's type is defined struct,
		// if it is ,
		// use errorCode as err and return struct,
		// otherwise use respBody string as error and return
		if errUnmarshal != nil || len(output.GetResponseMetadata().GetError().GetCode()) == 0 {
			if err == nil {
				err = errors.New(string(respBody))
			}
			return nil, status, err
		} else {
			return output, status, errors.New(output.GetResponseMetadata().GetError().GetCode())
		}
	}
	return output, status, nil
}

// UpdateUploadSpaceConfig
/*
 * @param *request.VodUpdateUploadSpaceConfigRequest
 * @return *response.VodUpdateUploadSpaceConfigResponse, int, error
 */
func (p *Vod) UpdateUploadSpaceConfig(req *request.VodUpdateUploadSpaceConfigRequest) (*response.VodUpdateUploadSpaceConfigResponse, int, error) {
	query := url.Values{}
	form := url.Values{}
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
		return nil, 0, err
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
				return nil, 0, e2
			}
			sv = string(v2)
		}
		query.Set(k, sv)
		form.Add(k, sv)
	}

	respBody, status, err := p.Json("UpdateUploadSpaceConfig", url.Values{}, jsonData)

	output := &response.VodUpdateUploadSpaceConfigResponse{}
	unmarshaler := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
	errUnmarshal := unmarshaler.Unmarshal(respBody, output)
	if err != nil || status != http.StatusOK {
		// if exist http err,check whether the respBody's type is defined struct,
		// if it is ,
		// return struct,
		// otherwise return nil body
		// if httpCode is not 200,check whether the respBody's type is defined struct,
		// if it is ,
		// use errorCode as err and return struct,
		// otherwise use respBody string as error and return
		if errUnmarshal != nil || len(output.GetResponseMetadata().GetError().GetCode()) == 0 {
			if err == nil {
				err = errors.New(string(respBody))
			}
			return nil, status, err
		} else {
			return output, status, errors.New(output.GetResponseMetadata().GetError().GetCode())
		}
	}
	return output, status, nil
}

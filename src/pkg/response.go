package pkg

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/emicklei/go-restful"
)

type Page struct {
	PageSize   int64  `json:"limit"`
	Offset     int64  `json:"offset"`
	Page       int64  `json:"page"`
	TotalCount int64  `json:"-"`
	Query      string `json:"-"`
	OrderBy    string `json:"order_by"`
	Sort       string `json:"sort"`
}

type Result struct {
	TotalCount  *int64      `json:"total_count,omitempty"`
	PageCount   *int64      `json:"page_count,omitempty"`
	CurrentPage *int64      `json:"current_page,omitempty"`
	PageSize    *int64      `json:"page_size,omitempty"`
	Results     interface{} `json:"result"`
	Code        int64       `json:"code"`
}

func NewResult(count int64, page *Page, results interface{}) Result {
	var result Result
	var pageCount int64
	if page != nil {
		result = Result{
			TotalCount:  &count,
			CurrentPage: &page.Page,
			PageSize:    &page.PageSize,
			PageCount:   &pageCount,
			Results:     results,
			Code:        0,
		}

		pc := count / page.PageSize
		result.PageCount = &pc
		if count%page.PageSize > 0 {
			*(result.PageCount) += 1
		}
		// 处理跳转至随机页，该情况下，currentPage 不为 1，开始模糊查询的问题
		if *result.CurrentPage > *result.PageCount {
			*result.CurrentPage = 1
		}
	} else {
		result = Result{
			Results: results,
			Code:    0,
		}
	}

	return result
}

type ResponseResult struct {
	Code   int64       `json:"code"`
	Result interface{} `json:"result"`
}

const (
	AuthFailed string = "401-10000"

	InvalidParams    string = "400-10001"
	InvalidUser      string = "400-10002"
	InvalidOperation string = "400-10003"

	HiError string = "400-20000"
)

func WriteResponse(resp *restful.Response, code string, result interface{}) {
	httpCode, res := NewResponseResult(code, result)
	_ = resp.WriteHeaderAndEntity(httpCode, res)
}

func NewResponseResult(code string, result interface{}) (int, ResponseResult) {
	codes := strings.Split(code, "-")
	httpCode, _ := strconv.Atoi(codes[0])
	statusCode, _ := strconv.ParseInt(codes[1], 10, 64)

	return httpCode, ResponseResult{
		Code:   statusCode,
		Result: fmt.Sprint(result),
	}
}

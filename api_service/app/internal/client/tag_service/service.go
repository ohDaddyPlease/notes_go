package tag_service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/fatih/structs"
	"gitlab.konstweb.ru/ow/arch/notes/api_service/internal/apperror"
	"gitlab.konstweb.ru/ow/arch/notes/pkg/logging"
	"gitlab.konstweb.ru/ow/arch/notes/pkg/rest"
	"net/http"
	"strings"
	"time"
)

var _ TagService = &client{}

type client struct {
	base     rest.BaseClient
	resource string
}

func NewService(baseURL string, resource string, logger logging.Logger) TagService {
	return &client{
		resource: resource,
		base: rest.BaseClient{
			BaseURL: baseURL,
			HTTPClient: &http.Client{
				Timeout: 10 * time.Second,
			},
			Logger: logger,
		},
	}
}

type TagService interface {
	GetOne(ctx context.Context, id int) ([]byte, error)
	GetMany(ctx context.Context, ids []int) ([]byte, error)
	Create(ctx context.Context, tag CreateTagDTO) (string, error)
	Update(ctx context.Context, uuid string, tag UpdateTagDTO) error
	Delete(ctx context.Context, id string) error
}

func (c *client) GetOne(ctx context.Context, id int) ([]byte, error) {
	var tags []byte

	uri, err := c.base.BuildURL(fmt.Sprintf("%s/%d", c.resource, id), nil)
	if err != nil {
		return tags, fmt.Errorf("failed to build URL. error: %v", err)
	}
	c.base.Logger.Tracef("url: %s", uri)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return tags, fmt.Errorf("failed to create new request due to error: %v", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req = req.WithContext(reqCtx)
	response, err := c.base.SendRequest(req)
	if err != nil {
		return tags, fmt.Errorf("failed to send request due to error: %v", err)
	}

	if response.IsOk {
		tags, err = response.ReadBody()
		if err != nil {
			return nil, fmt.Errorf("failed to read body")
		}
		return tags, nil
	}
	return nil, apperror.APIError(response.Error.ErrorCode, response.Error.Message, response.Error.DeveloperMessage)
}

func (c *client) GetMany(ctx context.Context, ids []int) ([]byte, error) {
	var tags []byte

	filters := []rest.FilterOptions{
		{
			Field:  "id",
			Values: strings.Split(strings.Trim(fmt.Sprint(ids), "[]"), " "),
		},
	}

	uri, err := c.base.BuildURL(c.resource, filters)
	if err != nil {
		return tags, fmt.Errorf("failed to build URL. error: %v", err)
	}
	c.base.Logger.Tracef("url: %s", uri)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return tags, fmt.Errorf("failed to create new request due to error: %v", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req = req.WithContext(reqCtx)
	response, err := c.base.SendRequest(req)
	if err != nil {
		return tags, fmt.Errorf("failed to send request due to error: %v", err)
	}

	if response.IsOk {
		tags, err = response.ReadBody()
		if err != nil {
			return nil, fmt.Errorf("failed to read body")
		}
		return tags, nil
	}
	return nil, apperror.APIError(response.Error.ErrorCode, response.Error.Message, response.Error.DeveloperMessage)
}

func (c *client) Create(ctx context.Context, tag CreateTagDTO) (string, error) {
	var tagUUID string

	uri, err := c.base.BuildURL(c.resource, nil)
	if err != nil {
		return tagUUID, fmt.Errorf("failed to build URL. error: %v", err)
	}
	c.base.Logger.Tracef("url: %s", uri)

	structs.DefaultTagName = "json"
	data := structs.Map(tag)

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return tagUUID, fmt.Errorf("failed to marshal dto")
	}

	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(dataBytes))
	if err != nil {
		return tagUUID, fmt.Errorf("failed to create new request due to error: %v", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req = req.WithContext(reqCtx)
	response, err := c.base.SendRequest(req)
	if err != nil {
		return tagUUID, fmt.Errorf("failed to send request due to error: %v", err)
	}

	if response.IsOk {
		tagURL, err := response.Location()
		if err != nil {
			return tagUUID, fmt.Errorf("failed to get Location header")
		}
		c.base.Logger.Tracef("Location: %s", tagURL.String())

		splitCategoryURL := strings.Split(tagURL.String(), "/")
		tagUUID = splitCategoryURL[len(splitCategoryURL)-1]
		return tagUUID, nil
	}
	return tagUUID, apperror.APIError(response.Error.ErrorCode, response.Error.Message, response.Error.DeveloperMessage)
}

func (c *client) Update(ctx context.Context, uuid string, tag UpdateTagDTO) error {
	uri, err := c.base.BuildURL(fmt.Sprintf("%s/%s", c.resource, uuid), nil)
	if err != nil {
		return fmt.Errorf("failed to build URL. error: %v", err)
	}
	c.base.Logger.Tracef("url: %s", uri)

	structs.DefaultTagName = "json"
	data := structs.Map(tag)

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal dto")
	}

	req, err := http.NewRequest("PATCH", uri, bytes.NewBuffer(dataBytes))
	if err != nil {
		return fmt.Errorf("failed to create new request due to error: %v", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req = req.WithContext(reqCtx)
	response, err := c.base.SendRequest(req)
	if err != nil {
		return fmt.Errorf("failed to send request due to error: %v", err)
	}

	if response.IsOk {
		return nil
	}
	return apperror.APIError(response.Error.ErrorCode, response.Error.Message, response.Error.DeveloperMessage)
}

func (c *client) Delete(ctx context.Context, id string) error {
	uri, err := c.base.BuildURL(fmt.Sprintf("%s/%s", c.resource, id), nil)
	if err != nil {
		return fmt.Errorf("failed to build URL. error: %v", err)
	}
	c.base.Logger.Tracef("url: %s", uri)

	req, err := http.NewRequest("DELETE", uri, nil)
	if err != nil {
		return fmt.Errorf("failed to create new request due to error: %v", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req = req.WithContext(reqCtx)
	response, err := c.base.SendRequest(req)
	if err != nil {
		return fmt.Errorf("failed to send request due to error: %v", err)
	}

	if response.IsOk {
		return nil
	}
	return apperror.APIError(response.Error.ErrorCode, response.Error.Message, response.Error.DeveloperMessage)
}

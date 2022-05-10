package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/iam1912/gemseries/gemdelayqueue/client/dqclient"
	"github.com/iam1912/gemseries/gemdelayqueue/consts"
	"github.com/iam1912/gemseries/gemdelayqueue/utils"
)

type Handler struct {
	Client *dqclient.Client
}

func New(client *dqclient.Client) Handler {
	return Handler{Client: client}
}

func (h Handler) Add(w http.ResponseWriter, r *http.Request) {
	cmd := CheckCommand(r)
	if cmd == "" || cmd != "add" {
		RenderFailureJSON(w, InvalidParam)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		RenderFailureJSON(w, err.Error())
		return
	}
	req := &AddRequest{}
	err = json.Unmarshal(body, &req)
	if err != nil {
		RenderFailureJSON(w, err.Error())
		return
	}
	if strings.HasPrefix(req.Topic, consts.JobPrefix) {
		RenderFailureJSON(w, "incorrct topic parameter format")
		return
	}
	err = h.Client.Add(context.Background(), req.ID, req.Topic, int64(req.Delay), int64(req.TTR), req.MaxTries, req.Body)
	if err != nil {
		RenderFailureJSON(w, fmt.Sprintf("%s is add failed:%s", utils.GetJobKey(req.Topic, req.ID), err.Error()))
		return
	}
}

func (h Handler) Pop(w http.ResponseWriter, r *http.Request) {
	cmd := CheckCommand(r)
	if cmd == "" || cmd != "pop" {
		RenderFailureJSON(w, InvalidParam)
		return
	}
	topic := r.FormValue("topic")
	if topic == "" {
		RenderFailureJSON(w, InvalidParam)
		return
	}
	job, err := h.Client.RPop(context.Background(), topic)
	if err != nil {
		RenderFailureJSON(w, fmt.Sprintf("%s is pop failed:%s", topic, err.Error()))
		return
	}
	RenderSuccessJSON(w, job.ID, job.Topic, job.Body)
}

func (h Handler) Finish(w http.ResponseWriter, r *http.Request) {
	cmd := CheckCommand(r)
	if cmd == "" || cmd != "finish" {
		RenderFailureJSON(w, InvalidParam)
		return
	}
	id := r.FormValue("id")
	if id == "" {
		RenderFailureJSON(w, InvalidParam)
		return
	}
	err := h.Client.Finish(context.Background(), id)
	if err != nil {
		RenderFailureJSON(w, fmt.Sprintf("%s is update state finish failed:%s", id, err.Error()))
		return
	}
}

func (h Handler) Delete(w http.ResponseWriter, r *http.Request) {
	cmd := CheckCommand(r)
	if cmd == "" || cmd != "delete" {
		RenderFailureJSON(w, InvalidParam)
		return
	}
	id := r.FormValue("id")
	if id == "" {
		RenderFailureJSON(w, InvalidParam)
		return
	}
	err := h.Client.Deleted(context.Background(), id)
	if err != nil {
		RenderFailureJSON(w, fmt.Sprintf("%s is update state deleted failed:%s", id, err.Error()))
		return
	}
}

func (h Handler) GetInfo(w http.ResponseWriter, r *http.Request) {
	cmd := CheckCommand(r)
	if cmd == "" || cmd != "info" {
		RenderFailureJSON(w, InvalidParam)
		return
	}
	id := r.FormValue("id")
	topic := r.FormValue("topic")
	if id == "" || topic == "" {
		RenderFailureJSON(w, InvalidParam)
		return
	}
	job, err := h.Client.GetJobInfo(context.Background(), topic, id)
	if err != nil {
		RenderFailureJSON(w, err.Error())
		return
	}
	RenderSuccessJSON(w, job.ID, job.Topic, job.Body)
}

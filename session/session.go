package session

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

//session接口
type Sessioner interface {
	Get(interface{}) interface{}
	Set(key,val interface{})
	Del(interface{})
	ID() string
	Reset() //删除session中所有元素
}


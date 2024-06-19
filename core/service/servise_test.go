package service

import (
	"capcha/adapters/cache"
	"capcha/config"
	"capcha/core/ports"
	"fmt"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

var testCache ports.Cache

func newService() (*Service, error) {
	cnf, err := config.GetConfig("")
	if err != nil {
		return nil, err
	}
	c, err := cache.NewRedisCache(cnf.Redis.Addr, cnf.Redis.Password)
	if err != nil {
		return nil, err
	}
	testCache = c
	return New(cnf.Service, c)
}

func TestGenerateCapcha(t *testing.T) {
	s, err := newService()
	require.NoError(t, err, "cannot create service")
	c, err := s.GenerateCapcha()
	require.NoError(t, err)
	fmt.Println("text val: ", c.TextVal)
	err = os.WriteFile("test.png", c.Image, 666)
	require.NoError(t, err, "cannot save file")
}

func TestGetCheckCapcha(t *testing.T) {
	s, err := newService()
	require.NoError(t, err, "cannot create service")
	c, err := s.NewCapcha()
	require.NoError(t, err)
	fmt.Println(c.Uid, c.Expires, c.ImgSrc)

	fmt.Println("uid:", c.Uid)
	validVal, err := testCache.Get(c.Uid)
	require.NoError(t, err)
	fmt.Println("valid val:", validVal)

	img, err := s.GetImage(c.Uid)
	require.NoError(t, err)
	err = os.WriteFile("C:\\Users\\79212\\GolandProjects\\capcha\\test_img\\"+c.Uid+".png", img, 666)
	require.NoError(t, err)

	isValid, err := s.CheckCapcha(c.Uid, "not valid value")
	require.NoError(t, err)
	require.Equal(t, false, isValid)

	isValid, err = s.CheckCapcha(c.Uid, validVal)
	require.NoError(t, err)
	require.Equal(t, true, isValid)

}

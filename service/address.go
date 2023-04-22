package service

import (
	"context"
	"strconv"
	"sync"

	logging "github.com/sirupsen/logrus"

	"mall/pkg/e"
	"mall/repository/db/dao"
	"mall/repository/db/model"
	"mall/types"
)

var AddressSrvIns *AddressSrv
var AddressSrvOnce sync.Once

type AddressSrv struct {
}

func GetAddressSrv() *AddressSrv {
	AddressSrvOnce.Do(func() {
		AddressSrvIns = &AddressSrv{}
	})
	return AddressSrvIns
}

func (s *AddressSrv) Create(ctx context.Context, req *types.AddressServiceReq, uId uint) (resp interface{}, err error) {
	code := e.SUCCESS
	addressDao := dao.NewAddressDao(ctx)
	address := &model.Address{
		UserID:  uId,
		Name:    req.Name,
		Phone:   req.Phone,
		Address: req.Address,
	}
	err = addressDao.CreateAddress(address)
	if err != nil {
		logging.Info(err)
		code = e.ErrorDatabase
		return types.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}, err
	}
	addressDao = dao.NewAddressDaoByDB(addressDao.DB)
	var addresses []*model.Address
	addresses, err = addressDao.ListAddressByUid(uId)
	if err != nil {
		logging.Info(err)
		code = e.ErrorDatabase
		return types.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}, err
	}
	return types.Response{
		Status: code,
		Data:   types.BuildAddresses(addresses),
		Msg:    e.GetMsg(code),
	}, nil
}

func (s *AddressSrv) Show(ctx context.Context, aId string) (resp interface{}, err error) {
	code := e.SUCCESS
	addressDao := dao.NewAddressDao(ctx)
	addressId, _ := strconv.Atoi(aId)
	address, err := addressDao.GetAddressByAid(uint(addressId))
	if err != nil {
		logging.Info(err)
		code = e.ErrorDatabase
		return types.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}, err
	}
	return types.Response{
		Status: code,
		Data:   types.BuildAddress(address),
		Msg:    e.GetMsg(code),
	}, nil
}

func (s *AddressSrv) List(ctx context.Context, uId uint) (resp interface{}, err error) {
	code := e.SUCCESS
	addressDao := dao.NewAddressDao(ctx)
	address, err := addressDao.ListAddressByUid(uId)
	if err != nil {
		logging.Info(err)
		code = e.ErrorDatabase
		return types.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}, err
	}
	return types.Response{
		Status: code,
		Data:   types.BuildAddresses(address),
		Msg:    e.GetMsg(code),
	}, nil
}

func (s *AddressSrv) Delete(ctx context.Context, aId, uId uint) (types.Response, error) {
	addressDao := dao.NewAddressDao(ctx)
	code := e.SUCCESS
	err := addressDao.DeleteAddressById(aId, uId)
	if err != nil {
		logging.Info(err)
		code = e.ErrorDatabase
		return types.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}, err
	}
	return types.Response{
		Status: code,
		Msg:    e.GetMsg(code),
	}, nil
}

func (s *AddressSrv) Update(ctx context.Context, req *types.AddressServiceReq, uid, aid uint) (types.Response, error) {
	code := e.SUCCESS

	addressDao := dao.NewAddressDao(ctx)
	address := &model.Address{
		UserID:  uid,
		Name:    req.Name,
		Phone:   req.Phone,
		Address: req.Address,
	}
	err := addressDao.UpdateAddressById(aid, address)
	addressDao = dao.NewAddressDaoByDB(addressDao.DB)
	var addresses []*model.Address
	addresses, err = addressDao.ListAddressByUid(uid)
	if err != nil {
		logging.Info(err)
		code = e.ErrorDatabase
		return types.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}, err
	}
	return types.Response{
		Status: code,
		Data:   types.BuildAddresses(addresses),
		Msg:    e.GetMsg(code),
	}, nil
}

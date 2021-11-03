package service

import (
	"FanOneMall/cache"
	"FanOneMall/model"
	"FanOneMall/pkg/e"
	logging "github.com/sirupsen/logrus"
	"FanOneMall/serializer"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"math/rand"
	"strconv"
	"time"
)

//Create
type CreateOrderService struct {
	UserID uint `form:"user_id" json:"user_id"`
	ProductID uint `form:"product_id" json:"product_id"`
	Num uint `form:"num" json:"num"`
	AddressID uint `form:"num" json:"address_id"`
	Money int `form:"money" json:"money"`
	BossID uint `form:"user_id" json:"boss_id"`
}

//Search
type ListOrdersService struct {
	Limit int `form:"limit" json:"limit"`
	Start int `form:"start" json:"start"`
	Type int `form:"type" json:"type"`
}

//Detail
type ShowOrderService struct {

}

type DeleteOrderService struct {
	UserID    uint `form:"user_id" json:"user_id"`
	ProductID uint `form:"product_id" json:"product_id"`
	OrderNum uint `form:"order_num" json:"order_num"`
}


func (service *CreateOrderService) Create() serializer.Response {
	order := model.Order{
		UserID:    service.UserID,
		ProductID: service.ProductID,
		BossID:    service.BossID,
		Num:       service.Num,
		Money: 	   service.Money,
		Type:      1,
	}
	address := model.Address{}
	code := e.SUCCESS
	if err := model.DB.First(&address,service.AddressID).Error;err!=nil{
		logging.Info(err)
		code = e.ErrorDatabase
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}
	}
	order.AddressID = address.ID
	number := fmt.Sprintf("%09v",rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(1000000000))
	productNum := strconv.Itoa(int(service.ProductID))
	userNum := strconv.Itoa(int(service.UserID))
	number = number+productNum+userNum
	orderNum,err := strconv.ParseUint(number,10,64)
	if err != nil {
		logging.Info(err)
		code = e.ERROR
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}
	}
	order.OrderNum = orderNum
	err = model.DB.Create(&order).Error
	if err != nil {
		logging.Info(err)
		code = e.ErrorDatabase
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}
	}
	//订单号存入Redis中，设置过期时间
	data:=redis.Z{
		Score:  float64(time.Now().Unix()) + 15*time.Minute.Seconds(),
		Member: orderNum,
	}
	cache.RedisClient.ZAdd("3",data)
	return serializer.Response{
		Status: code,
		Msg:    e.GetMsg(code),
	}
}


func (service *ListOrdersService) List(id string) serializer.Response{
	var orders []model.Order
	total := 0
	code := e.SUCCESS
	if service.Limit == 0{
		service.Limit =5
	}
	if service.Type == 0{
		if err := model.DB.Model(&orders).Where("user_id=?",id).Count(&total).Error;err!=nil{
			logging.Info(err)
			code = e.ErrorDatabase
			return serializer.Response{
				Status: code,
				Msg:    e.GetMsg(code),
				Error:  err.Error(),
			}
		}
		if err := model.DB.Where("user_id=?",id).Limit(service.Limit).Offset(service.Start).Order("created_at desc").Find(&orders).Error;err!=nil{
			logging.Info(err)
			code = e.ErrorDatabase
			return serializer.Response{
				Status: code,
				Msg:    e.GetMsg(code),
				Error:  err.Error(),
			}
		}
	}else {
		if err := model.DB.Model(&orders).Where("user_id=? AND type = ?" ,id,service.Type).Count(&total).Error;err!=nil{
			logging.Info(err)
			code = e.ErrorDatabase
			return serializer.Response{
				Status: code,
				Msg:    e.GetMsg(code),
				Error:  err.Error(),
			}
		}
		if err := model.DB.Where("user_id=? AND type=?",id,service.Type).Limit(service.Limit).Offset(service.Start).Order("created_at desc").Find(&orders).Error; err != nil {
				logging.Info(err)
				code = e.ErrorDatabase
				return serializer.Response{
					Status: code,
					Msg:    e.GetMsg(code),
					Error:  err.Error(),
				}
			}
		}
	return serializer.BuildListResponse(serializer.BuildOrders(orders),uint(total))
}



func (service *ShowOrderService) Show(num string) serializer.Response {
	var order model.Order
	var product model.Product
	var address model.Address
	code := e.SUCCESS

	if err := model.DB.Where("order_num=?",num).First(&order).Error;err!=nil{
		logging.Info(err)
		code = e.ErrorDatabase
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
		}
	}
	model.DB.Where("id = ?",order.AddressID).First(&address)
	if err := model.DB.Where("id=?",order.ProductID).First(&product).Error;err!=nil{
		if gorm.IsRecordNotFoundError(err){
			logging.Info(err)
			code = e.ErrorNotExistProduct
			return serializer.Response{
				Status: code,
				Msg:    e.GetMsg(code),
			}
		}
		logging.Info(err)
		code = e.ErrorDatabase
		return serializer.Response{
			Status:code,
			Msg:e.GetMsg(code),
		}
	}
	return serializer.Response{
		Status:code,
		Msg:e.GetMsg(code),
		Data:serializer.BuildOrder(order,product,address),
	}
}

func (service *DeleteOrderService) Delete() serializer.Response {
	var order model.Order
	code := e.SUCCESS
	err := model.DB.Where("order_num=?", service.OrderNum).Find(&order).Error
	if err != nil {
		logging.Info(err)
		code = e.ErrorDatabase
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}
	}
	err = model.DB.Delete(&order).Error
	if err != nil {
		logging.Info(err)
		code = e.ErrorDatabase
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}
	}
	return serializer.Response{
		Status: code,
		Msg:    e.GetMsg(code),
	}
}

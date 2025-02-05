# hummingbird-sdk-go

#### 介绍
蜂鸟物联网平台官方Go语言SDK。

#### 方法

## GetDBClient

获取数据库连接
```go
func (d *DriverService) GetDBClient() *gorm.DB
```

## Online
设备在线

```go
func (d *DriverService) Online (deviceId string) error
```

## Offline
设备离线

```go
func (d *DriverService) Offline (deviceId string) error 
```


## GetConnectStatus
获取设备连接状态
```go
func (d *DriverService) GetConnectStatus(deviceId string) (commons.DeviceConnectStatus, error)
```


## CreateDevice
创建设备

```go
func (d *DriverService) CreateDevice(device model.AddDevice) (model.Device, error) 
```

## UpdateDevice
更新设备

```go
func (d *DriverService) UpdateDevice(device model.UpdateDevice) (model.Device, error) 
```

## DeleteDevice
删除设备

```go
func (d *DriverService) DeleteDevice(deviceId string) error
```


## GetDeviceList
获取所有的设备

```go
func (d *DriverService) GetDeviceList() map[string]model.Device 
```

## GetDeviceById
通过设备id获取设备详情
```go
func (d *DriverService) GetDeviceById(deviceId string) (model.Device, bool) 
```

## ProductList
获取当前实例下的所有产品
```go
func (d *DriverService) ProductList() map[string]model.Product 
```


## GetProductById
根据产品id获取产品信息
```go
func (d *DriverService) GetProductById(productId string) (model.Product, bool) 
```

## GetProductProperties
根据产品id获取产品所有属性信息
```go
func (d *DriverService) GetProductProperties(productId string) (map[string]model.Property, bool) 
```

## GetProductPropertyByCode
根据产品id与code获取属性信息
```go
func (d *DriverService) GetProductPropertyByCode(productId, code string) (model.Property, bool) 
```

## GetProductEvents
根据产品id获取产品所有事件信息
```go
func (d *DriverService) GetProductEvents(productId string) (map[string]model.Event, bool) 
```

## GetProductEventByCode
根据产品id与code获取事件信息
```go
func (d *DriverService) GetProductEventByCode(productId, code string) (model.Event, bool) 
```

## GetPropertyServices
根据产品id获取产品所有服务信息
```go
func (d *DriverService) GetPropertyServices(productId string) (map[string]model.Service, bool) 
```

## GetProductServiceByCode
根据产品id与code获取服务信息
```go
func (d *DriverService) GetProductServiceByCode(productId, code string) (model.Service, bool) 
```

## PropertyReport
物模型属性上报 如果data参数中的Sys.Ack设置为1，则该方法会同步阻塞等待云端返回结果。
```go
func (d *DriverService) PropertyReport(deviceId string, data model.PropertyReport) (model.CommonResponse, error) 
```

## EventReport
物模型事件上报
```go
func (d *DriverService) EventReport(deviceId string, data model.EventReport) (model.CommonResponse, error) 
```

## BatchReport
设备批量上报属性和事件
```go
func (d *DriverService) BatchReport(deviceId string, data model.BatchReport) (model.CommonResponse, error) 
```


## PropertySetResponse
设备属性下发响应
```go
func (d *DriverService) PropertySetResponse(deviceId string, data model.CommonResponse) error 
```

## PropertyGetResponse
设备属性查询响应
```go
func (d *DriverService) PropertyGetResponse(deviceId string, data model.PropertyGetResponse) error
```

## ServiceExecuteResponse
设备动作执行响应
```go
func (d *DriverService) ServiceExecuteResponse(deviceId string, data model.ServiceExecuteResponse) error 
```

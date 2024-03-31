/*******************************************************************************
 * Copyright 2017.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

package model

import (
    "github.com/winc-link/edge-driver-proto/driverproduct"
    "github.com/winc-link/hummingbird-sdk-go/commons"
)

type (
    Product struct {
        //CreateAt     time.Time
        Id           string
        Name         string
        Key          string
        Description  string
        NodeType     commons.ProductNodeType
        DataFormat   string
        Platform     commons.IotPlatform
        NetType      commons.ProductNetType
        ProtocolType string
        Properties   []Property //属性
        Events       []Event    //事件
        Services     []Service  //服务
    }
    
    Service struct {
        ProductId   string
        Name        string
        Code        string
        Required    bool
        CallType    string
        Description string
        InputData   []InputData
        OutputData  []OutputData
    }
    
    TypeSpec struct {
        Type  string
        Specs string
    }
    
    Event struct {
        ProductId   string
        Name        string
        Code        string
        Required    bool
        Type        string
        Description string
        OutputData  []OutputData
    }
    
    Property struct {
        ProductId   string
        Name        string
        Code        string
        Description string
        Required    bool
        AccessMode  string
        TypeSpec    TypeSpec
        Extend      string
    }
    
    OutputData struct {
        Code     string
        Name     string
        TypeSpec TypeSpec
    }
    
    InputData struct {
        Code     string
        Name     string
        TypeSpec TypeSpec
    }
)

func TransformProductModel(p *driverproduct.Product) Product {
    return Product{
        Id:           p.GetId(),
        Name:         p.GetName(),
        Key:          p.GetKey(),
        Description:  p.GetDescription(),
        NodeType:     commons.TransformRpcNodeTypeToModel(p.NodeType),
        Platform:     commons.TransformRpcPlatformToModel(p.Platform),
        NetType:      commons.TransformRpcNetTypeToModel(p.NetType),
        ProtocolType: p.GetProtocolType(),
        Properties:   propertyModels(p.GetProperties()),
        Events:       eventModels(p.GetEvents()),
        Services:     serviceModels(p.GetActions()),
    }
}

func propertyModels(p []*driverproduct.Properties) []Property {
    rets := make([]Property, 0, len(p))
    for i := range p {
        rets = append(rets, Property{
            ProductId:   p[i].GetProductId(),
            Name:        p[i].GetName(),
            Code:        p[i].GetCode(),
            Description: p[i].GetDescription(),
            Required:    p[i].GetRequired(),
            AccessMode:  p[i].GetAccessMode(),
            TypeSpec:    TransformTypeSpecModel(p[i].GetTypeSpec()),
            Extend:      p[i].GetExtend(),
        })
    }
    return rets
}

func eventModels(e []*driverproduct.Events) []Event {
    rets := make([]Event, 0, len(e))
    for i := range e {
        rets = append(rets, Event{
            ProductId:   e[i].GetProductId(),
            Name:        e[i].GetName(),
            Code:        e[i].GetCode(),
            Required:    e[i].GetRequired(),
            Type:        e[i].GetType(),
            Description: e[i].GetDescription(),
            OutputData:  TransformOutputData(e[i].GetOutputParams()),
        })
    }
    return rets
}

func serviceModels(as []*driverproduct.Actions) []Service {
    rets := make([]Service, 0, len(as))
    for i := range as {
        rets = append(rets, Service{
            ProductId:   as[i].GetProductId(),
            Name:        as[i].GetName(),
            Code:        as[i].GetCode(),
            Required:    as[i].GetRequired(),
            CallType:    as[i].CallType,
            Description: as[i].GetDescription(),
            InputData:   TransformInputData(as[i].GetInputParams()),
            OutputData:  TransformOutputData(as[i].GetOutputParams()),
        })
    }
    return rets
}

func TransformInputData(params []*driverproduct.InputParams) []InputData {
    rets := make([]InputData, 0, len(params))
    for i := range params {
        rets = append(rets, InputData{
            Code:     params[i].GetCode(),
            Name:     params[i].GetName(),
            TypeSpec: TransformTypeSpecModel(params[i].GetTypeSpec()),
        })
    }
    return rets
}

func TransformOutputData(params []*driverproduct.OutputParams) []OutputData {
    rets := make([]OutputData, 0, len(params))
    for i := range params {
        rets = append(rets, OutputData{
            Code:     params[i].GetCode(),
            Name:     params[i].GetName(),
            TypeSpec: TransformTypeSpecModel(params[i].GetTypeSpec()),
        })
    }
    return rets
}

func TransformTypeSpecModel(spec *driverproduct.TypeSpec) TypeSpec {
    return TypeSpec{
        Type:  spec.GetType(),
        Specs: spec.GetSpecs(),
    }
}

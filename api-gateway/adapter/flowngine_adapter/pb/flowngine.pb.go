// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v6.31.0
// source: flowngine.proto

package pb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Transfer status enum
type TransferStatus int32

const (
	TransferStatus_TRANSFER_STATUS_UNSPECIFIED TransferStatus = 0
	TransferStatus_TRANSFER_STATUS_PENDING     TransferStatus = 1
	TransferStatus_TRANSFER_STATUS_PROCESSING  TransferStatus = 2
	TransferStatus_TRANSFER_STATUS_COMPLETED   TransferStatus = 3
	TransferStatus_TRANSFER_STATUS_FAILED      TransferStatus = 4
	TransferStatus_TRANSFER_STATUS_COMPENSATED TransferStatus = 5
	TransferStatus_TRANSFER_STATUS_CANCELLED   TransferStatus = 6
)

// Enum value maps for TransferStatus.
var (
	TransferStatus_name = map[int32]string{
		0: "TRANSFER_STATUS_UNSPECIFIED",
		1: "TRANSFER_STATUS_PENDING",
		2: "TRANSFER_STATUS_PROCESSING",
		3: "TRANSFER_STATUS_COMPLETED",
		4: "TRANSFER_STATUS_FAILED",
		5: "TRANSFER_STATUS_COMPENSATED",
		6: "TRANSFER_STATUS_CANCELLED",
	}
	TransferStatus_value = map[string]int32{
		"TRANSFER_STATUS_UNSPECIFIED": 0,
		"TRANSFER_STATUS_PENDING":     1,
		"TRANSFER_STATUS_PROCESSING":  2,
		"TRANSFER_STATUS_COMPLETED":   3,
		"TRANSFER_STATUS_FAILED":      4,
		"TRANSFER_STATUS_COMPENSATED": 5,
		"TRANSFER_STATUS_CANCELLED":   6,
	}
)

func (x TransferStatus) Enum() *TransferStatus {
	p := new(TransferStatus)
	*p = x
	return p
}

func (x TransferStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (TransferStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_flowngine_proto_enumTypes[0].Descriptor()
}

func (TransferStatus) Type() protoreflect.EnumType {
	return &file_flowngine_proto_enumTypes[0]
}

func (x TransferStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use TransferStatus.Descriptor instead.
func (TransferStatus) EnumDescriptor() ([]byte, []int) {
	return file_flowngine_proto_rawDescGZIP(), []int{0}
}

// Transfer request message
type ExecuteTransferRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	FromAccount   string                 `protobuf:"bytes,1,opt,name=from_account,json=fromAccount,proto3" json:"from_account,omitempty"`
	ToAccount     string                 `protobuf:"bytes,2,opt,name=to_account,json=toAccount,proto3" json:"to_account,omitempty"`
	Amount        int64                  `protobuf:"varint,3,opt,name=amount,proto3" json:"amount,omitempty"`
	Currency      string                 `protobuf:"bytes,4,opt,name=currency,proto3" json:"currency,omitempty"`
	Description   string                 `protobuf:"bytes,5,opt,name=description,proto3" json:"description,omitempty"`
	ReferenceId   string                 `protobuf:"bytes,6,opt,name=reference_id,json=referenceId,proto3" json:"reference_id,omitempty"`
	RequestId     string                 `protobuf:"bytes,7,opt,name=request_id,json=requestId,proto3" json:"request_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ExecuteTransferRequest) Reset() {
	*x = ExecuteTransferRequest{}
	mi := &file_flowngine_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ExecuteTransferRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExecuteTransferRequest) ProtoMessage() {}

func (x *ExecuteTransferRequest) ProtoReflect() protoreflect.Message {
	mi := &file_flowngine_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExecuteTransferRequest.ProtoReflect.Descriptor instead.
func (*ExecuteTransferRequest) Descriptor() ([]byte, []int) {
	return file_flowngine_proto_rawDescGZIP(), []int{0}
}

func (x *ExecuteTransferRequest) GetFromAccount() string {
	if x != nil {
		return x.FromAccount
	}
	return ""
}

func (x *ExecuteTransferRequest) GetToAccount() string {
	if x != nil {
		return x.ToAccount
	}
	return ""
}

func (x *ExecuteTransferRequest) GetAmount() int64 {
	if x != nil {
		return x.Amount
	}
	return 0
}

func (x *ExecuteTransferRequest) GetCurrency() string {
	if x != nil {
		return x.Currency
	}
	return ""
}

func (x *ExecuteTransferRequest) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *ExecuteTransferRequest) GetReferenceId() string {
	if x != nil {
		return x.ReferenceId
	}
	return ""
}

func (x *ExecuteTransferRequest) GetRequestId() string {
	if x != nil {
		return x.RequestId
	}
	return ""
}

// Transfer response message
type ExecuteTransferResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	TransactionId string                 `protobuf:"bytes,1,opt,name=transaction_id,json=transactionId,proto3" json:"transaction_id,omitempty"`
	Status        TransferStatus         `protobuf:"varint,2,opt,name=status,proto3,enum=pb.TransferStatus" json:"status,omitempty"`
	WorkflowId    string                 `protobuf:"bytes,3,opt,name=workflow_id,json=workflowId,proto3" json:"workflow_id,omitempty"`
	RunId         string                 `protobuf:"bytes,4,opt,name=run_id,json=runId,proto3" json:"run_id,omitempty"`
	CreatedAt     *timestamppb.Timestamp `protobuf:"bytes,5,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ExecuteTransferResponse) Reset() {
	*x = ExecuteTransferResponse{}
	mi := &file_flowngine_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ExecuteTransferResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExecuteTransferResponse) ProtoMessage() {}

func (x *ExecuteTransferResponse) ProtoReflect() protoreflect.Message {
	mi := &file_flowngine_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExecuteTransferResponse.ProtoReflect.Descriptor instead.
func (*ExecuteTransferResponse) Descriptor() ([]byte, []int) {
	return file_flowngine_proto_rawDescGZIP(), []int{1}
}

func (x *ExecuteTransferResponse) GetTransactionId() string {
	if x != nil {
		return x.TransactionId
	}
	return ""
}

func (x *ExecuteTransferResponse) GetStatus() TransferStatus {
	if x != nil {
		return x.Status
	}
	return TransferStatus_TRANSFER_STATUS_UNSPECIFIED
}

func (x *ExecuteTransferResponse) GetWorkflowId() string {
	if x != nil {
		return x.WorkflowId
	}
	return ""
}

func (x *ExecuteTransferResponse) GetRunId() string {
	if x != nil {
		return x.RunId
	}
	return ""
}

func (x *ExecuteTransferResponse) GetCreatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.CreatedAt
	}
	return nil
}

// Status request message
type GetTransferStatusRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	TransactionId string                 `protobuf:"bytes,1,opt,name=transaction_id,json=transactionId,proto3" json:"transaction_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetTransferStatusRequest) Reset() {
	*x = GetTransferStatusRequest{}
	mi := &file_flowngine_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetTransferStatusRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetTransferStatusRequest) ProtoMessage() {}

func (x *GetTransferStatusRequest) ProtoReflect() protoreflect.Message {
	mi := &file_flowngine_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetTransferStatusRequest.ProtoReflect.Descriptor instead.
func (*GetTransferStatusRequest) Descriptor() ([]byte, []int) {
	return file_flowngine_proto_rawDescGZIP(), []int{2}
}

func (x *GetTransferStatusRequest) GetTransactionId() string {
	if x != nil {
		return x.TransactionId
	}
	return ""
}

// Status response message
type GetTransferStatusResponse struct {
	state             protoimpl.MessageState `protogen:"open.v1"`
	TransactionId     string                 `protobuf:"bytes,1,opt,name=transaction_id,json=transactionId,proto3" json:"transaction_id,omitempty"`
	Status            TransferStatus         `protobuf:"varint,2,opt,name=status,proto3,enum=pb.TransferStatus" json:"status,omitempty"`
	FromAccount       string                 `protobuf:"bytes,3,opt,name=from_account,json=fromAccount,proto3" json:"from_account,omitempty"`
	ToAccount         string                 `protobuf:"bytes,4,opt,name=to_account,json=toAccount,proto3" json:"to_account,omitempty"`
	Amount            int64                  `protobuf:"varint,5,opt,name=amount,proto3" json:"amount,omitempty"`
	Currency          string                 `protobuf:"bytes,6,opt,name=currency,proto3" json:"currency,omitempty"`
	Description       string                 `protobuf:"bytes,7,opt,name=description,proto3" json:"description,omitempty"`
	ReferenceId       string                 `protobuf:"bytes,8,opt,name=reference_id,json=referenceId,proto3" json:"reference_id,omitempty"`
	CreatedAt         *timestamppb.Timestamp `protobuf:"bytes,9,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	CompletedAt       *timestamppb.Timestamp `protobuf:"bytes,10,opt,name=completed_at,json=completedAt,proto3" json:"completed_at,omitempty"`
	WorkflowExecution *WorkflowExecution     `protobuf:"bytes,11,opt,name=workflow_execution,json=workflowExecution,proto3" json:"workflow_execution,omitempty"`
	ErrorMessage      string                 `protobuf:"bytes,12,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
	unknownFields     protoimpl.UnknownFields
	sizeCache         protoimpl.SizeCache
}

func (x *GetTransferStatusResponse) Reset() {
	*x = GetTransferStatusResponse{}
	mi := &file_flowngine_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetTransferStatusResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetTransferStatusResponse) ProtoMessage() {}

func (x *GetTransferStatusResponse) ProtoReflect() protoreflect.Message {
	mi := &file_flowngine_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetTransferStatusResponse.ProtoReflect.Descriptor instead.
func (*GetTransferStatusResponse) Descriptor() ([]byte, []int) {
	return file_flowngine_proto_rawDescGZIP(), []int{3}
}

func (x *GetTransferStatusResponse) GetTransactionId() string {
	if x != nil {
		return x.TransactionId
	}
	return ""
}

func (x *GetTransferStatusResponse) GetStatus() TransferStatus {
	if x != nil {
		return x.Status
	}
	return TransferStatus_TRANSFER_STATUS_UNSPECIFIED
}

func (x *GetTransferStatusResponse) GetFromAccount() string {
	if x != nil {
		return x.FromAccount
	}
	return ""
}

func (x *GetTransferStatusResponse) GetToAccount() string {
	if x != nil {
		return x.ToAccount
	}
	return ""
}

func (x *GetTransferStatusResponse) GetAmount() int64 {
	if x != nil {
		return x.Amount
	}
	return 0
}

func (x *GetTransferStatusResponse) GetCurrency() string {
	if x != nil {
		return x.Currency
	}
	return ""
}

func (x *GetTransferStatusResponse) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *GetTransferStatusResponse) GetReferenceId() string {
	if x != nil {
		return x.ReferenceId
	}
	return ""
}

func (x *GetTransferStatusResponse) GetCreatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.CreatedAt
	}
	return nil
}

func (x *GetTransferStatusResponse) GetCompletedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.CompletedAt
	}
	return nil
}

func (x *GetTransferStatusResponse) GetWorkflowExecution() *WorkflowExecution {
	if x != nil {
		return x.WorkflowExecution
	}
	return nil
}

func (x *GetTransferStatusResponse) GetErrorMessage() string {
	if x != nil {
		return x.ErrorMessage
	}
	return ""
}

// Cancel request message
type CancelTransferRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	TransactionId string                 `protobuf:"bytes,1,opt,name=transaction_id,json=transactionId,proto3" json:"transaction_id,omitempty"`
	Reason        string                 `protobuf:"bytes,2,opt,name=reason,proto3" json:"reason,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CancelTransferRequest) Reset() {
	*x = CancelTransferRequest{}
	mi := &file_flowngine_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CancelTransferRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CancelTransferRequest) ProtoMessage() {}

func (x *CancelTransferRequest) ProtoReflect() protoreflect.Message {
	mi := &file_flowngine_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CancelTransferRequest.ProtoReflect.Descriptor instead.
func (*CancelTransferRequest) Descriptor() ([]byte, []int) {
	return file_flowngine_proto_rawDescGZIP(), []int{4}
}

func (x *CancelTransferRequest) GetTransactionId() string {
	if x != nil {
		return x.TransactionId
	}
	return ""
}

func (x *CancelTransferRequest) GetReason() string {
	if x != nil {
		return x.Reason
	}
	return ""
}

// Cancel response message
type CancelTransferResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Success       bool                   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	Message       string                 `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CancelTransferResponse) Reset() {
	*x = CancelTransferResponse{}
	mi := &file_flowngine_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CancelTransferResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CancelTransferResponse) ProtoMessage() {}

func (x *CancelTransferResponse) ProtoReflect() protoreflect.Message {
	mi := &file_flowngine_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CancelTransferResponse.ProtoReflect.Descriptor instead.
func (*CancelTransferResponse) Descriptor() ([]byte, []int) {
	return file_flowngine_proto_rawDescGZIP(), []int{5}
}

func (x *CancelTransferResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

func (x *CancelTransferResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

// Workflow execution details
type WorkflowExecution struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	WorkflowId    string                 `protobuf:"bytes,1,opt,name=workflow_id,json=workflowId,proto3" json:"workflow_id,omitempty"`
	RunId         string                 `protobuf:"bytes,2,opt,name=run_id,json=runId,proto3" json:"run_id,omitempty"`
	Status        string                 `protobuf:"bytes,3,opt,name=status,proto3" json:"status,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *WorkflowExecution) Reset() {
	*x = WorkflowExecution{}
	mi := &file_flowngine_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *WorkflowExecution) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WorkflowExecution) ProtoMessage() {}

func (x *WorkflowExecution) ProtoReflect() protoreflect.Message {
	mi := &file_flowngine_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WorkflowExecution.ProtoReflect.Descriptor instead.
func (*WorkflowExecution) Descriptor() ([]byte, []int) {
	return file_flowngine_proto_rawDescGZIP(), []int{6}
}

func (x *WorkflowExecution) GetWorkflowId() string {
	if x != nil {
		return x.WorkflowId
	}
	return ""
}

func (x *WorkflowExecution) GetRunId() string {
	if x != nil {
		return x.RunId
	}
	return ""
}

func (x *WorkflowExecution) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

var File_flowngine_proto protoreflect.FileDescriptor

const file_flowngine_proto_rawDesc = "" +
	"\n" +
	"\x0fflowngine.proto\x12\x02pb\x1a\x1fgoogle/protobuf/timestamp.proto\"\xf2\x01\n" +
	"\x16ExecuteTransferRequest\x12!\n" +
	"\ffrom_account\x18\x01 \x01(\tR\vfromAccount\x12\x1d\n" +
	"\n" +
	"to_account\x18\x02 \x01(\tR\ttoAccount\x12\x16\n" +
	"\x06amount\x18\x03 \x01(\x03R\x06amount\x12\x1a\n" +
	"\bcurrency\x18\x04 \x01(\tR\bcurrency\x12 \n" +
	"\vdescription\x18\x05 \x01(\tR\vdescription\x12!\n" +
	"\freference_id\x18\x06 \x01(\tR\vreferenceId\x12\x1d\n" +
	"\n" +
	"request_id\x18\a \x01(\tR\trequestId\"\xdf\x01\n" +
	"\x17ExecuteTransferResponse\x12%\n" +
	"\x0etransaction_id\x18\x01 \x01(\tR\rtransactionId\x12*\n" +
	"\x06status\x18\x02 \x01(\x0e2\x12.pb.TransferStatusR\x06status\x12\x1f\n" +
	"\vworkflow_id\x18\x03 \x01(\tR\n" +
	"workflowId\x12\x15\n" +
	"\x06run_id\x18\x04 \x01(\tR\x05runId\x129\n" +
	"\n" +
	"created_at\x18\x05 \x01(\v2\x1a.google.protobuf.TimestampR\tcreatedAt\"A\n" +
	"\x18GetTransferStatusRequest\x12%\n" +
	"\x0etransaction_id\x18\x01 \x01(\tR\rtransactionId\"\x8e\x04\n" +
	"\x19GetTransferStatusResponse\x12%\n" +
	"\x0etransaction_id\x18\x01 \x01(\tR\rtransactionId\x12*\n" +
	"\x06status\x18\x02 \x01(\x0e2\x12.pb.TransferStatusR\x06status\x12!\n" +
	"\ffrom_account\x18\x03 \x01(\tR\vfromAccount\x12\x1d\n" +
	"\n" +
	"to_account\x18\x04 \x01(\tR\ttoAccount\x12\x16\n" +
	"\x06amount\x18\x05 \x01(\x03R\x06amount\x12\x1a\n" +
	"\bcurrency\x18\x06 \x01(\tR\bcurrency\x12 \n" +
	"\vdescription\x18\a \x01(\tR\vdescription\x12!\n" +
	"\freference_id\x18\b \x01(\tR\vreferenceId\x129\n" +
	"\n" +
	"created_at\x18\t \x01(\v2\x1a.google.protobuf.TimestampR\tcreatedAt\x12=\n" +
	"\fcompleted_at\x18\n" +
	" \x01(\v2\x1a.google.protobuf.TimestampR\vcompletedAt\x12D\n" +
	"\x12workflow_execution\x18\v \x01(\v2\x15.pb.WorkflowExecutionR\x11workflowExecution\x12#\n" +
	"\rerror_message\x18\f \x01(\tR\ferrorMessage\"V\n" +
	"\x15CancelTransferRequest\x12%\n" +
	"\x0etransaction_id\x18\x01 \x01(\tR\rtransactionId\x12\x16\n" +
	"\x06reason\x18\x02 \x01(\tR\x06reason\"L\n" +
	"\x16CancelTransferResponse\x12\x18\n" +
	"\asuccess\x18\x01 \x01(\bR\asuccess\x12\x18\n" +
	"\amessage\x18\x02 \x01(\tR\amessage\"c\n" +
	"\x11WorkflowExecution\x12\x1f\n" +
	"\vworkflow_id\x18\x01 \x01(\tR\n" +
	"workflowId\x12\x15\n" +
	"\x06run_id\x18\x02 \x01(\tR\x05runId\x12\x16\n" +
	"\x06status\x18\x03 \x01(\tR\x06status*\xe9\x01\n" +
	"\x0eTransferStatus\x12\x1f\n" +
	"\x1bTRANSFER_STATUS_UNSPECIFIED\x10\x00\x12\x1b\n" +
	"\x17TRANSFER_STATUS_PENDING\x10\x01\x12\x1e\n" +
	"\x1aTRANSFER_STATUS_PROCESSING\x10\x02\x12\x1d\n" +
	"\x19TRANSFER_STATUS_COMPLETED\x10\x03\x12\x1a\n" +
	"\x16TRANSFER_STATUS_FAILED\x10\x04\x12\x1f\n" +
	"\x1bTRANSFER_STATUS_COMPENSATED\x10\x05\x12\x1d\n" +
	"\x19TRANSFER_STATUS_CANCELLED\x10\x062\xf3\x01\n" +
	"\n" +
	"FlowEngine\x12J\n" +
	"\x0fExecuteTransfer\x12\x1a.pb.ExecuteTransferRequest\x1a\x1b.pb.ExecuteTransferResponse\x12P\n" +
	"\x11GetTransferStatus\x12\x1c.pb.GetTransferStatusRequest\x1a\x1d.pb.GetTransferStatusResponse\x12G\n" +
	"\x0eCancelTransfer\x12\x19.pb.CancelTransferRequest\x1a\x1a.pb.CancelTransferResponseB\x06Z\x04./pbb\x06proto3"

var (
	file_flowngine_proto_rawDescOnce sync.Once
	file_flowngine_proto_rawDescData []byte
)

func file_flowngine_proto_rawDescGZIP() []byte {
	file_flowngine_proto_rawDescOnce.Do(func() {
		file_flowngine_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_flowngine_proto_rawDesc), len(file_flowngine_proto_rawDesc)))
	})
	return file_flowngine_proto_rawDescData
}

var file_flowngine_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_flowngine_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_flowngine_proto_goTypes = []any{
	(TransferStatus)(0),               // 0: pb.TransferStatus
	(*ExecuteTransferRequest)(nil),    // 1: pb.ExecuteTransferRequest
	(*ExecuteTransferResponse)(nil),   // 2: pb.ExecuteTransferResponse
	(*GetTransferStatusRequest)(nil),  // 3: pb.GetTransferStatusRequest
	(*GetTransferStatusResponse)(nil), // 4: pb.GetTransferStatusResponse
	(*CancelTransferRequest)(nil),     // 5: pb.CancelTransferRequest
	(*CancelTransferResponse)(nil),    // 6: pb.CancelTransferResponse
	(*WorkflowExecution)(nil),         // 7: pb.WorkflowExecution
	(*timestamppb.Timestamp)(nil),     // 8: google.protobuf.Timestamp
}
var file_flowngine_proto_depIdxs = []int32{
	0, // 0: pb.ExecuteTransferResponse.status:type_name -> pb.TransferStatus
	8, // 1: pb.ExecuteTransferResponse.created_at:type_name -> google.protobuf.Timestamp
	0, // 2: pb.GetTransferStatusResponse.status:type_name -> pb.TransferStatus
	8, // 3: pb.GetTransferStatusResponse.created_at:type_name -> google.protobuf.Timestamp
	8, // 4: pb.GetTransferStatusResponse.completed_at:type_name -> google.protobuf.Timestamp
	7, // 5: pb.GetTransferStatusResponse.workflow_execution:type_name -> pb.WorkflowExecution
	1, // 6: pb.FlowEngine.ExecuteTransfer:input_type -> pb.ExecuteTransferRequest
	3, // 7: pb.FlowEngine.GetTransferStatus:input_type -> pb.GetTransferStatusRequest
	5, // 8: pb.FlowEngine.CancelTransfer:input_type -> pb.CancelTransferRequest
	2, // 9: pb.FlowEngine.ExecuteTransfer:output_type -> pb.ExecuteTransferResponse
	4, // 10: pb.FlowEngine.GetTransferStatus:output_type -> pb.GetTransferStatusResponse
	6, // 11: pb.FlowEngine.CancelTransfer:output_type -> pb.CancelTransferResponse
	9, // [9:12] is the sub-list for method output_type
	6, // [6:9] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_flowngine_proto_init() }
func file_flowngine_proto_init() {
	if File_flowngine_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_flowngine_proto_rawDesc), len(file_flowngine_proto_rawDesc)),
			NumEnums:      1,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_flowngine_proto_goTypes,
		DependencyIndexes: file_flowngine_proto_depIdxs,
		EnumInfos:         file_flowngine_proto_enumTypes,
		MessageInfos:      file_flowngine_proto_msgTypes,
	}.Build()
	File_flowngine_proto = out.File
	file_flowngine_proto_goTypes = nil
	file_flowngine_proto_depIdxs = nil
}

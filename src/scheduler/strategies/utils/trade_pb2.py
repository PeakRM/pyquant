# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# NO CHECKED-IN PROTOBUF GENCODE
# source: trade.proto
# Protobuf Python Version: 5.28.1
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import runtime_version as _runtime_version
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
_runtime_version.ValidateProtobufRuntimeVersion(
    _runtime_version.Domain.PUBLIC,
    5,
    28,
    1,
    '',
    'trade.proto'
)
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x0btrade.proto\x12\x05trade\"\xa8\x01\n\x05Trade\x12\x15\n\rstrategy_name\x18\x01 \x01(\t\x12\x13\n\x0b\x63ontract_id\x18\x02 \x01(\x05\x12\x10\n\x08\x65xchange\x18\x03 \x01(\t\x12\x0e\n\x06symbol\x18\x04 \x01(\t\x12\x0c\n\x04side\x18\x05 \x01(\t\x12\x10\n\x08quantity\x18\x06 \x01(\t\x12\x12\n\norder_type\x18\x07 \x01(\t\x12\x0e\n\x06\x62roker\x18\x08 \x01(\t\x12\r\n\x05price\x18\t \x01(\t\"\x1f\n\rTradeResponse\x12\x0e\n\x06status\x18\x01 \x01(\t2?\n\x0cTradeService\x12/\n\tSendTrade\x12\x0c.trade.Trade\x1a\x14.trade.TradeResponseB\x13Z\x11scheduler/tradepbb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'trade_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  _globals['DESCRIPTOR']._loaded_options = None
  _globals['DESCRIPTOR']._serialized_options = b'Z\021scheduler/tradepb'
  _globals['_TRADE']._serialized_start=23
  _globals['_TRADE']._serialized_end=191
  _globals['_TRADERESPONSE']._serialized_start=193
  _globals['_TRADERESPONSE']._serialized_end=224
  _globals['_TRADESERVICE']._serialized_start=226
  _globals['_TRADESERVICE']._serialized_end=289
# @@protoc_insertion_point(module_scope)

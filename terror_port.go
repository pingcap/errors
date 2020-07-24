// Copyright 2020 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

// For reviewers: This file is only use for porting this terror to `parser/terror`.

package errors

import "fmt"

const (
	defaultMySQLErrorCode = 1105
	// DefaultMySQLState is default state of the mySQL
	defaultMySQLState = "HY000"

	ErrDupKey = 1022

	ErrOutofMemory     = 1037
	ErrOutOfSortMemory = 1038

	ErrConCount = 1040

	ErrBadHost               = 1042
	ErrHandshake             = 1043
	ErrDBaccessDenied        = 1044
	ErrAccessDenied          = 1045
	ErrNoDB                  = 1046
	ErrUnknownCom            = 1047
	ErrBadNull               = 1048
	ErrBadDB                 = 1049
	ErrTableExists           = 1050
	ErrBadTable              = 1051
	ErrNonUniq               = 1052
	ErrServerShutdown        = 1053
	ErrBadField              = 1054
	ErrFieldNotInGroupBy     = 1055
	ErrWrongGroupField       = 1056
	ErrWrongSumSelect        = 1057
	ErrWrongValueCount       = 1058
	ErrTooLongIdent          = 1059
	ErrDupFieldName          = 1060
	ErrDupKeyName            = 1061
	ErrDupEntry              = 1062
	ErrWrongFieldSpec        = 1063
	ErrParse                 = 1064
	ErrEmptyQuery            = 1065
	ErrNonuniqTable          = 1066
	ErrInvalidDefault        = 1067
	ErrMultiplePriKey        = 1068
	ErrTooManyKeys           = 1069
	ErrTooManyKeyParts       = 1070
	ErrTooLongKey            = 1071
	ErrKeyColumnDoesNotExits = 1072
	ErrBlobUsedAsKey         = 1073
	ErrTooBigFieldlength     = 1074
	ErrWrongAutoKey          = 1075

	ErrForcingClose          = 1080
	ErrIpsock                = 1081
	ErrNoSuchIndex           = 1082
	ErrWrongFieldTerminators = 1083
	ErrBlobsAndNoTerminated  = 1084

	ErrCantRemoveAllFields = 1090
	ErrCantDropFieldOrKey  = 1091

	ErrBlobCantHaveDefault = 1101
	ErrWrongDBName         = 1102
	ErrWrongTableName      = 1103
	ErrTooBigSelect        = 1104

	ErrUnknownProcedure           = 1106
	ErrWrongParamcountToProcedure = 1107

	ErrUnknownTable        = 1109
	ErrFieldSpecifiedTwice = 1110

	ErrUnsupportedExtension = 1112
	ErrTableMustHaveColumns = 1113

	ErrUnknownCharacterSet = 1115

	ErrTooBigRowsize = 1118

	ErrWrongOuterJoin    = 1120
	ErrNullColumnInIndex = 1121

	ErrPasswordAnonymousUser = 1131
	ErrPasswordNotAllowed    = 1132
	ErrPasswordNoMatch       = 1133

	ErrWrongValueCountOnRow = 1136

	ErrInvalidUseOfNull        = 1138
	ErrRegexp                  = 1139
	ErrMixOfGroupFuncAndFields = 1140
	ErrNonexistingGrant        = 1141
	ErrTableaccessDenied       = 1142
	ErrColumnaccessDenied      = 1143
	ErrIllegalGrantForTable    = 1144
	ErrGrantWrongHostOrUser    = 1145
	ErrNoSuchTable             = 1146
	ErrNonexistingTableGrant   = 1147
	ErrNotAllowedCommand       = 1148
	ErrSyntax                  = 1149

	ErrAbortingConnection           = 1152
	ErrNetPacketTooLarge            = 1153
	ErrNetReadErrorFromPipe         = 1154
	ErrNetFcntl                     = 1155
	ErrNetPacketsOutOfOrder         = 1156
	ErrNetUncompress                = 1157
	ErrNetRead                      = 1158
	ErrNetReadInterrupted           = 1159
	ErrNetErrorOnWrite              = 1160
	ErrNetWriteInterrupted          = 1161
	ErrTooLongString                = 1162
	ErrTableCantHandleBlob          = 1163
	ErrTableCantHandleAutoIncrement = 1164

	ErrWrongColumnName = 1166
	ErrWrongKeyColumn  = 1167

	ErrDupUnique            = 1169
	ErrBlobKeyWithoutLength = 1170
	ErrPrimaryCantHaveNull  = 1171
	ErrTooManyRows          = 1172
	ErrRequiresPrimaryKey   = 1173

	ErrKeyDoesNotExist               = 1176
	ErrCheckNoSuchTable              = 1177
	ErrCheckNotImplemented           = 1178
	ErrCantDoThisDuringAnTransaction = 1179

	ErrNewAbortingConnection = 1184

	ErrMasterNetRead  = 1189
	ErrMasterNetWrite = 1190

	ErrTooManyUserConnections = 1203

	ErrReadOnlyTransaction = 1207

	ErrNoPermissionToCreateUser = 1211

	ErrLockDeadlock = 1213

	ErrNoReferencedRow = 1216
	ErrRowIsReferenced = 1217
	ErrConnectToMaster = 1218

	ErrWrongNumberOfColumnsInSelect = 1222

	ErrUserLimitReached     = 1226
	ErrSpecificAccessDenied = 1227

	ErrNoDefault        = 1230
	ErrWrongValueForVar = 1231
	ErrWrongTypeForVar  = 1232

	ErrCantUseOptionHere = 1234
	ErrNotSupportedYet   = 1235

	ErrWrongFkDef = 1239

	ErrOperandColumns = 1241
	ErrSubqueryNo1Row = 1242

	ErrIllegalReference         = 1247
	ErrDerivedMustHaveAlias     = 1248
	ErrSelectReduced            = 1249
	ErrTablenameNotAllowedHere  = 1250
	ErrNotSupportedAuthMode     = 1251
	ErrSpatialCantHaveNull      = 1252
	ErrCollationCharsetMismatch = 1253

	ErrWarnTooFewRecords  = 1261
	ErrWarnTooManyRecords = 1262
	ErrWarnNullToNotnull  = 1263
	ErrWarnDataOutOfRange = 1264
	WarnDataTruncated     = 1265

	ErrWrongNameForIndex   = 1280
	ErrWrongNameForCatalog = 1281

	ErrUnknownStorageEngine = 1286

	ErrTruncatedWrongValue = 1292

	ErrSpNoRecursiveCreate = 1303
	ErrSpAlreadyExists     = 1304
	ErrSpDoesNotExist      = 1305

	ErrSpLilabelMismatch             = 1308
	ErrSpLabelRedefine               = 1309
	ErrSpLabelMismatch               = 1310
	ErrSpUninitVar                   = 1311
	ErrSpBadselect                   = 1312
	ErrSpBadreturn                   = 1313
	ErrSpBadstatement                = 1314
	ErrUpdateLogDeprecatedIgnored    = 1315
	ErrUpdateLogDeprecatedTranslated = 1316
	ErrQueryInterrupted              = 1317
	ErrSpWrongNoOfArgs               = 1318
	ErrSpCondMismatch                = 1319
	ErrSpNoreturn                    = 1320
	ErrSpNoreturnend                 = 1321
	ErrSpBadCursorQuery              = 1322
	ErrSpBadCursorSelect             = 1323
	ErrSpCursorMismatch              = 1324
	ErrSpCursorAlreadyOpen           = 1325
	ErrSpCursorNotOpen               = 1326
	ErrSpUndeclaredVar               = 1327

	ErrSpFetchNoData = 1329
	ErrSpDupParam    = 1330
	ErrSpDupVar      = 1331
	ErrSpDupCond     = 1332
	ErrSpDupCurs     = 1333

	ErrSpSubselectNyi          = 1335
	ErrStmtNotAllowedInSfOrTrg = 1336
	ErrSpVarcondAfterCurshndlr = 1337
	ErrSpCursorAfterHandler    = 1338
	ErrSpCaseNotFound          = 1339

	ErrDivisionByZero = 1365

	ErrIllegalValueForType = 1367

	ErrProcaccessDenied = 1370

	ErrXaerNota             = 1397
	ErrXaerInval            = 1398
	ErrXaerRmfail           = 1399
	ErrXaerOutside          = 1400
	ErrXaerRmerr            = 1401
	ErrXaRbrollback         = 1402
	ErrNonexistingProcGrant = 1403

	ErrDataTooLong   = 1406
	ErrSpBadSQLstate = 1407

	ErrCantCreateUserWithGrant = 1410

	ErrSpDupHandler             = 1413
	ErrSpNotVarArg              = 1414
	ErrSpNoRetset               = 1415
	ErrCantCreateGeometryObject = 1416

	ErrTooBigScale     = 1425
	ErrTooBigPrecision = 1426
	ErrMBiggerThanD    = 1427

	ErrTooLongBody = 1437

	ErrTooBigDisplaywidth       = 1439
	ErrXaerDupid                = 1440
	ErrDatetimeFunctionOverflow = 1441

	ErrRowIsReferenced2 = 1451
	ErrNoReferencedRow2 = 1452
	ErrSpBadVarShadow   = 1453

	ErrSpWrongName = 1458

	ErrSpNoAggregate               = 1460
	ErrMaxPreparedStmtCountReached = 1461

	ErrNonGroupingFieldUsed = 1463

	ErrForeignDuplicateKeyOldUnused = 1557

	ErrCantChangeTxCharacteristics = 1568

	ErrWrongParamcountToNativeFct = 1582
	ErrWrongParametersToNativeFct = 1583
	ErrWrongParametersToStoredFct = 1584

	ErrDupEntryWithKeyName = 1586

	ErrXaRbtimeout  = 1613
	ErrXaRbdeadlock = 1614

	ErrFuncInexistentNameCollision = 1630

	ErrDupSignalSet                 = 1641
	ErrSignalWarn                   = 1642
	ErrSignalNotFound               = 1643
	ErrSignalException              = 1644
	ErrResignalWithoutActiveHandler = 1645

	ErrSpatialMustHaveGeomCol = 1687

	ErrDataOutOfRange = 1690

	ErrAccessDeniedNoPassword = 1698

	ErrTruncateIllegalFk = 1701

	ErrDaInvalidConditionNumber = 1758

	ErrForeignDuplicateKeyWithChildInfo    = 1761
	ErrForeignDuplicateKeyWithoutChildInfo = 1762

	ErrCantExecuteInReadOnlyTransaction = 1792

	ErrAlterOperationNotSupported       = 1845
	ErrAlterOperationNotSupportedReason = 1846

	ErrDupUnknownInIndex = 1859

	ErrInvalidJSONData = 3069

	ErrBadGeneratedColumn           = 3105
	ErrUnsupportedOnGeneratedColumn = 3106
	ErrGeneratedColumnNonPrior      = 3107
	ErrDependentByGeneratedColumn   = 3108

	ErrInvalidJSONText = 3140
	ErrInvalidJSONPath = 3143

	ErrInvalidJSONPathWildcard = 3149

	ErrJSONUsedAsKey       = 3152
	ErrJSONDocumentNULLKey = 3158

	ErrInvalidJSONPathArrayCell = 3165
)

// mySQLState maps error code to MySQL SQLSTATE value.
// The values are taken from ANSI SQL and ODBC and are more standardized.
var mySQLState = map[uint16]string{
	ErrDupKey:                              "23000",
	ErrOutofMemory:                         "HY001",
	ErrOutOfSortMemory:                     "HY001",
	ErrConCount:                            "08004",
	ErrBadHost:                             "08S01",
	ErrHandshake:                           "08S01",
	ErrDBaccessDenied:                      "42000",
	ErrAccessDenied:                        "28000",
	ErrNoDB:                                "3D000",
	ErrUnknownCom:                          "08S01",
	ErrBadNull:                             "23000",
	ErrBadDB:                               "42000",
	ErrTableExists:                         "42S01",
	ErrBadTable:                            "42S02",
	ErrNonUniq:                             "23000",
	ErrServerShutdown:                      "08S01",
	ErrBadField:                            "42S22",
	ErrFieldNotInGroupBy:                   "42000",
	ErrWrongSumSelect:                      "42000",
	ErrWrongGroupField:                     "42000",
	ErrWrongValueCount:                     "21S01",
	ErrTooLongIdent:                        "42000",
	ErrDupFieldName:                        "42S21",
	ErrDupKeyName:                          "42000",
	ErrDupEntry:                            "23000",
	ErrWrongFieldSpec:                      "42000",
	ErrParse:                               "42000",
	ErrEmptyQuery:                          "42000",
	ErrNonuniqTable:                        "42000",
	ErrInvalidDefault:                      "42000",
	ErrMultiplePriKey:                      "42000",
	ErrTooManyKeys:                         "42000",
	ErrTooManyKeyParts:                     "42000",
	ErrTooLongKey:                          "42000",
	ErrKeyColumnDoesNotExits:               "42000",
	ErrBlobUsedAsKey:                       "42000",
	ErrTooBigFieldlength:                   "42000",
	ErrWrongAutoKey:                        "42000",
	ErrForcingClose:                        "08S01",
	ErrIpsock:                              "08S01",
	ErrNoSuchIndex:                         "42S12",
	ErrWrongFieldTerminators:               "42000",
	ErrBlobsAndNoTerminated:                "42000",
	ErrCantRemoveAllFields:                 "42000",
	ErrCantDropFieldOrKey:                  "42000",
	ErrBlobCantHaveDefault:                 "42000",
	ErrWrongDBName:                         "42000",
	ErrWrongTableName:                      "42000",
	ErrTooBigSelect:                        "42000",
	ErrUnknownProcedure:                    "42000",
	ErrWrongParamcountToProcedure:          "42000",
	ErrUnknownTable:                        "42S02",
	ErrFieldSpecifiedTwice:                 "42000",
	ErrUnsupportedExtension:                "42000",
	ErrTableMustHaveColumns:                "42000",
	ErrUnknownCharacterSet:                 "42000",
	ErrTooBigRowsize:                       "42000",
	ErrWrongOuterJoin:                      "42000",
	ErrNullColumnInIndex:                   "42000",
	ErrPasswordAnonymousUser:               "42000",
	ErrPasswordNotAllowed:                  "42000",
	ErrPasswordNoMatch:                     "42000",
	ErrWrongValueCountOnRow:                "21S01",
	ErrInvalidUseOfNull:                    "22004",
	ErrRegexp:                              "42000",
	ErrMixOfGroupFuncAndFields:             "42000",
	ErrNonexistingGrant:                    "42000",
	ErrTableaccessDenied:                   "42000",
	ErrColumnaccessDenied:                  "42000",
	ErrIllegalGrantForTable:                "42000",
	ErrGrantWrongHostOrUser:                "42000",
	ErrNoSuchTable:                         "42S02",
	ErrNonexistingTableGrant:               "42000",
	ErrNotAllowedCommand:                   "42000",
	ErrSyntax:                              "42000",
	ErrAbortingConnection:                  "08S01",
	ErrNetPacketTooLarge:                   "08S01",
	ErrNetReadErrorFromPipe:                "08S01",
	ErrNetFcntl:                            "08S01",
	ErrNetPacketsOutOfOrder:                "08S01",
	ErrNetUncompress:                       "08S01",
	ErrNetRead:                             "08S01",
	ErrNetReadInterrupted:                  "08S01",
	ErrNetErrorOnWrite:                     "08S01",
	ErrNetWriteInterrupted:                 "08S01",
	ErrTooLongString:                       "42000",
	ErrTableCantHandleBlob:                 "42000",
	ErrTableCantHandleAutoIncrement:        "42000",
	ErrWrongColumnName:                     "42000",
	ErrWrongKeyColumn:                      "42000",
	ErrDupUnique:                           "23000",
	ErrBlobKeyWithoutLength:                "42000",
	ErrPrimaryCantHaveNull:                 "42000",
	ErrTooManyRows:                         "42000",
	ErrRequiresPrimaryKey:                  "42000",
	ErrKeyDoesNotExist:                     "42000",
	ErrCheckNoSuchTable:                    "42000",
	ErrCheckNotImplemented:                 "42000",
	ErrCantDoThisDuringAnTransaction:       "25000",
	ErrNewAbortingConnection:               "08S01",
	ErrMasterNetRead:                       "08S01",
	ErrMasterNetWrite:                      "08S01",
	ErrTooManyUserConnections:              "42000",
	ErrReadOnlyTransaction:                 "25000",
	ErrNoPermissionToCreateUser:            "42000",
	ErrLockDeadlock:                        "40001",
	ErrNoReferencedRow:                     "23000",
	ErrRowIsReferenced:                     "23000",
	ErrConnectToMaster:                     "08S01",
	ErrWrongNumberOfColumnsInSelect:        "21000",
	ErrUserLimitReached:                    "42000",
	ErrSpecificAccessDenied:                "42000",
	ErrNoDefault:                           "42000",
	ErrWrongValueForVar:                    "42000",
	ErrWrongTypeForVar:                     "42000",
	ErrCantUseOptionHere:                   "42000",
	ErrNotSupportedYet:                     "42000",
	ErrWrongFkDef:                          "42000",
	ErrOperandColumns:                      "21000",
	ErrSubqueryNo1Row:                      "21000",
	ErrIllegalReference:                    "42S22",
	ErrDerivedMustHaveAlias:                "42000",
	ErrSelectReduced:                       "01000",
	ErrTablenameNotAllowedHere:             "42000",
	ErrNotSupportedAuthMode:                "08004",
	ErrSpatialCantHaveNull:                 "42000",
	ErrCollationCharsetMismatch:            "42000",
	ErrWarnTooFewRecords:                   "01000",
	ErrWarnTooManyRecords:                  "01000",
	ErrWarnNullToNotnull:                   "22004",
	ErrWarnDataOutOfRange:                  "22003",
	WarnDataTruncated:                      "01000",
	ErrWrongNameForIndex:                   "42000",
	ErrWrongNameForCatalog:                 "42000",
	ErrUnknownStorageEngine:                "42000",
	ErrTruncatedWrongValue:                 "22007",
	ErrSpNoRecursiveCreate:                 "2F003",
	ErrSpAlreadyExists:                     "42000",
	ErrSpDoesNotExist:                      "42000",
	ErrSpLilabelMismatch:                   "42000",
	ErrSpLabelRedefine:                     "42000",
	ErrSpLabelMismatch:                     "42000",
	ErrSpUninitVar:                         "01000",
	ErrSpBadselect:                         "0A000",
	ErrSpBadreturn:                         "42000",
	ErrSpBadstatement:                      "0A000",
	ErrUpdateLogDeprecatedIgnored:          "42000",
	ErrUpdateLogDeprecatedTranslated:       "42000",
	ErrQueryInterrupted:                    "70100",
	ErrSpWrongNoOfArgs:                     "42000",
	ErrSpCondMismatch:                      "42000",
	ErrSpNoreturn:                          "42000",
	ErrSpNoreturnend:                       "2F005",
	ErrSpBadCursorQuery:                    "42000",
	ErrSpBadCursorSelect:                   "42000",
	ErrSpCursorMismatch:                    "42000",
	ErrSpCursorAlreadyOpen:                 "24000",
	ErrSpCursorNotOpen:                     "24000",
	ErrSpUndeclaredVar:                     "42000",
	ErrSpFetchNoData:                       "02000",
	ErrSpDupParam:                          "42000",
	ErrSpDupVar:                            "42000",
	ErrSpDupCond:                           "42000",
	ErrSpDupCurs:                           "42000",
	ErrSpSubselectNyi:                      "0A000",
	ErrStmtNotAllowedInSfOrTrg:             "0A000",
	ErrSpVarcondAfterCurshndlr:             "42000",
	ErrSpCursorAfterHandler:                "42000",
	ErrSpCaseNotFound:                      "20000",
	ErrDivisionByZero:                      "22012",
	ErrIllegalValueForType:                 "22007",
	ErrProcaccessDenied:                    "42000",
	ErrXaerNota:                            "XAE04",
	ErrXaerInval:                           "XAE05",
	ErrXaerRmfail:                          "XAE07",
	ErrXaerOutside:                         "XAE09",
	ErrXaerRmerr:                           "XAE03",
	ErrXaRbrollback:                        "XA100",
	ErrNonexistingProcGrant:                "42000",
	ErrDataTooLong:                         "22001",
	ErrSpBadSQLstate:                       "42000",
	ErrCantCreateUserWithGrant:             "42000",
	ErrSpDupHandler:                        "42000",
	ErrSpNotVarArg:                         "42000",
	ErrSpNoRetset:                          "0A000",
	ErrCantCreateGeometryObject:            "22003",
	ErrTooBigScale:                         "42000",
	ErrTooBigPrecision:                     "42000",
	ErrMBiggerThanD:                        "42000",
	ErrTooLongBody:                         "42000",
	ErrTooBigDisplaywidth:                  "42000",
	ErrXaerDupid:                           "XAE08",
	ErrDatetimeFunctionOverflow:            "22008",
	ErrRowIsReferenced2:                    "23000",
	ErrNoReferencedRow2:                    "23000",
	ErrSpBadVarShadow:                      "42000",
	ErrSpWrongName:                         "42000",
	ErrSpNoAggregate:                       "42000",
	ErrMaxPreparedStmtCountReached:         "42000",
	ErrNonGroupingFieldUsed:                "42000",
	ErrForeignDuplicateKeyOldUnused:        "23000",
	ErrCantChangeTxCharacteristics:         "25001",
	ErrWrongParamcountToNativeFct:          "42000",
	ErrWrongParametersToNativeFct:          "42000",
	ErrWrongParametersToStoredFct:          "42000",
	ErrDupEntryWithKeyName:                 "23000",
	ErrXaRbtimeout:                         "XA106",
	ErrXaRbdeadlock:                        "XA102",
	ErrFuncInexistentNameCollision:         "42000",
	ErrDupSignalSet:                        "42000",
	ErrSignalWarn:                          "01000",
	ErrSignalNotFound:                      "02000",
	ErrSignalException:                     "HY000",
	ErrResignalWithoutActiveHandler:        "0K000",
	ErrSpatialMustHaveGeomCol:              "42000",
	ErrDataOutOfRange:                      "22003",
	ErrAccessDeniedNoPassword:              "28000",
	ErrTruncateIllegalFk:                   "42000",
	ErrDaInvalidConditionNumber:            "35000",
	ErrForeignDuplicateKeyWithChildInfo:    "23000",
	ErrForeignDuplicateKeyWithoutChildInfo: "23000",
	ErrCantExecuteInReadOnlyTransaction:    "25006",
	ErrAlterOperationNotSupported:          "0A000",
	ErrAlterOperationNotSupportedReason:    "0A000",
	ErrDupUnknownInIndex:                   "23000",
	ErrBadGeneratedColumn:                  "HY000",
	ErrUnsupportedOnGeneratedColumn:        "HY000",
	ErrGeneratedColumnNonPrior:             "HY000",
	ErrDependentByGeneratedColumn:          "HY000",
	ErrInvalidJSONText:                     "22032",
	ErrInvalidJSONPath:                     "42000",
	ErrInvalidJSONData:                     "22032",
	ErrInvalidJSONPathWildcard:             "42000",
	ErrJSONUsedAsKey:                       "42000",
	ErrJSONDocumentNULLKey:                 "22032",
	ErrInvalidJSONPathArrayCell:            "42000",
}

// SQLError is polyfill for mysql.SQLError
type SQLError struct {
	Code    uint16
	Message string
	State   string
}

func (e *Error) getMySQLCode() uint16 {
	if e.class != nil && e.class.registry != nil {
		return uint16(e.Code())
	}
	return defaultMySQLErrorCode
}

func getStateForCode(code uint16) string {
	if state, ok := mySQLState[code]; ok {
		return state
	}
	return defaultMySQLState
}

func (e *Error) ToSQLError() *SQLError {
	return &SQLError{
		Code:    e.getMySQLCode(),
		Message: e.getMsg(),
		State:   getStateForCode(e.getMySQLCode()),
	}
}

// Error prints errors, with a formatted string.
func (e *SQLError) Error() string {
	return fmt.Sprintf("ERROR %d (%s): %s", e.Code, e.State, e.Message)
}

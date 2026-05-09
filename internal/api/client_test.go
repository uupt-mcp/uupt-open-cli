package api

import (
	"crypto/md5"
	"fmt"
	"strings"
	"testing"
)

func TestSignatureCompatibility_OrderPrice(t *testing.T) {
	bizParams := map[string]interface{}{
		"fromAddress":    "郑州市金水区花园路",
		"toAddress":      "郑州市二七区大学路",
		"sendType":       "SEND",
		"cityName":       "郑州市",
		"specialChannel": 2,
	}
	appSecret := "47888e576b454a23a1f4a34abaaac95d"
	timestamp := 1700000000

	sign := GenerateSign(bizParams, appSecret, timestamp)

	if len(sign) != 32 {
		t.Errorf("签名长度应为32，实际为%d", len(sign))
	}
	if sign != strings.ToUpper(sign) {
		t.Errorf("签名应为大写")
	}
}

func TestBizJsonSerialization(t *testing.T) {
	bizParams := map[string]interface{}{
		"fromAddress": "郑州市金水区花园路",
		"toAddress":   "郑州市二七区大学路",
	}

	bizJson, err := jsonAPI.MarshalToString(bizParams)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(bizJson, "\\u") {
		t.Errorf("中文字符被转义了: %s", bizJson)
	}

	if strings.Contains(bizJson, ": ") || strings.Contains(bizJson, ", ") {
		t.Errorf("JSON应为紧凑格式: %s", bizJson)
	}
}

func TestSignatureMatchesPython(t *testing.T) {
	// Python代码验证:
	//   import json, hashlib
	//   biz = {"cityName":"郑州市","fromAddress":"郑州市金水区花园路","sendType":"SEND","specialChannel":2,"toAddress":"郑州市二七区大学路"}
	//   biz_json = json.dumps(biz, ensure_ascii=False, separators=(",",":"), sort_keys=True)
	//   sign_str = biz_json + "47888e576b454a23a1f4a34abaaac95d" + "1700000000"
	//   sign = hashlib.md5(sign_str.encode("utf-8")).hexdigest().upper()

	bizParams := map[string]interface{}{
		"fromAddress":    "郑州市金水区花园路",
		"toAddress":      "郑州市二七区大学路",
		"sendType":       "SEND",
		"cityName":       "郑州市",
		"specialChannel": 2,
	}
	appSecret := "47888e576b454a23a1f4a34abaaac95d"
	timestamp := 1700000000

	sign := GenerateSign(bizParams, appSecret, timestamp)

	bizJson, _ := jsonAPI.MarshalToString(bizParams)
	t.Logf("bizJson: %s", bizJson)
	t.Logf("sign: %s", sign)

	// json-iterator SortMapKeys=true 按键名字母序排列
	// 预期键序: cityName, fromAddress, sendType, specialChannel, toAddress
	expectedBizJson := `{"cityName":"郑州市","fromAddress":"郑州市金水区花园路","sendType":"SEND","specialChannel":2,"toAddress":"郑州市二七区大学路"}`
	if bizJson != expectedBizJson {
		t.Errorf("bizJson不匹配\n期望: %s\n实际: %s", expectedBizJson, bizJson)
	}

	// 计算预期签名
	signStr := expectedBizJson + appSecret + "1700000000"
	expectedHash := md5.Sum([]byte(signStr))
	expectedSign := strings.ToUpper(fmt.Sprintf("%x", expectedHash))

	if sign != expectedSign {
		t.Errorf("签名不匹配\n期望: %s\n实际: %s", expectedSign, sign)
	}
}

func TestIntegerSerialization(t *testing.T) {
	bizParams := map[string]interface{}{
		"specialChannel": 2,
		"count":          10,
	}
	bizJson, _ := jsonAPI.MarshalToString(bizParams)
	if strings.Contains(bizJson, ".0") {
		t.Errorf("整数不应序列化为浮点: %s", bizJson)
	}
}

func TestSignatureWithEmptyBiz(t *testing.T) {
	bizParams := map[string]interface{}{}
	appSecret := "test-secret"
	timestamp := 1700000000

	sign := GenerateSign(bizParams, appSecret, timestamp)

	if len(sign) != 32 {
		t.Errorf("空biz签名长度应为32，实际为%d", len(sign))
	}
	if sign != strings.ToUpper(sign) {
		t.Errorf("签名应为大写")
	}
}

func TestJsonSortedKeys(t *testing.T) {
	bizParams := map[string]interface{}{
		"z_last":  1,
		"a_first": 2,
		"m_mid":   3,
	}
	bizJson, _ := jsonAPI.MarshalToString(bizParams)
	expected := `{"a_first":2,"m_mid":3,"z_last":1}`
	if bizJson != expected {
		t.Errorf("键排序不正确\n期望: %s\n实际: %s", expected, bizJson)
	}
}

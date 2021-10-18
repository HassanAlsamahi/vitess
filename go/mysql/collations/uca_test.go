/*
Copyright 2021 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package collations

import (
	"bytes"
	"strings"
	"testing"
	"unicode/utf8"

	"vitess.io/vitess/go/mysql/collations/internal/uca"
)

func testcollation(t testing.TB, name string) Collation {
	t.Helper()
	coll := LookupByName(name)
	if coll == nil {
		t.Fatalf("missing collation: %s", name)
	}
	return coll
}

func TestWeightsForSpace(t *testing.T) {
	for _, coll := range All() {
		var actual, expected uint16
		switch coll := coll.(type) {
		case *Collation_uca_legacy:
			actual = coll.uca.WeightForSpace()
			if strings.Contains(coll.name, "_520_") {
				expected = 0x20A
			} else {
				expected = 0x209
			}
		case *Collation_utf8mb4_uca_0900:
			actual = coll.uca.WeightForSpace()
			expected = 0x209
		default:
			continue
		}
		if actual != expected {
			t.Errorf("expected Weight(' ') == 0x%X, got 0x%X", expected, actual)
		}
	}
}

func TestKanaSensitivity(t *testing.T) {
	const Kana1 = "の東京ノ"
	const Kana2 = "ノ東京の"

	var cases = []struct {
		collation string
		equal     bool
	}{
		{"utf8mb4_0900_as_cs", false},
		{"utf8mb4_ja_0900_as_cs", true},
		{"utf8mb4_ja_0900_as_cs_ks", false},
	}

	for _, tc := range cases {
		t.Run(tc.collation, func(t *testing.T) {
			collation := testcollation(t, tc.collation)
			equal := collation.Collate([]byte(Kana1), []byte(Kana2), false) == 0
			if equal != tc.equal {
				t.Errorf("expected %q == %q to be %v", Kana1, Kana2, tc.equal)
			}
		})
	}
}

func TestContractions(t *testing.T) {
	var cases = []struct {
		collation string
		inputs    []string
		expected  []byte
	}{
		{
			collation: "utf8mb4_es_trad_0900_ai_ci",
			inputs:    []string{"ch", "CH", "Ch"},
			expected:  []byte{0x1c, 0x7a, 0x54, 0xa5},
		},
		{
			collation: "utf8mb4_es_trad_0900_as_cs",
			inputs:    []string{"Ch"},
			expected:  []byte{0x1c, 0x7a, 0x54, 0xa5, 0x0, 0x0, 0x0, 0x20, 0x0, 0x0, 0x0, 0x8, 0x0, 0x21},
		},
	}

	for _, tc := range cases {
		t.Run(tc.collation, func(t *testing.T) {
			coll := testcollation(t, tc.collation)

			for _, in := range tc.inputs {
				weightString := coll.WeightString(nil, []byte(in), 0)
				if !bytes.Equal(weightString, tc.expected) {
					t.Errorf("weight_string(%q) = %#v (expected %#v)", in, weightString, tc.expected)
				}
			}
		})
	}
}

func TestReplacementCharacter(t *testing.T) {
	var cases = []struct {
		collation string
		expected  []byte
	}{
		{"utf8mb4_0900_ai_ci", []byte{0xff, 0xfd}},
		{"utf8mb4_0900_as_ci", []byte{0xff, 0xfd, 0x0, 0x0, 0x0, 0x20}},
		{"utf8mb4_0900_as_cs", []byte{0xff, 0xfd, 0x0, 0x0, 0x0, 0x20, 0x0, 0x0, 0x0, 0x2}},
	}

	for _, tc := range cases {
		t.Run(tc.collation, func(t *testing.T) {
			coll := testcollation(t, tc.collation)
			weightString := coll.WeightString(nil, []byte(string(utf8.RuneError)), 0)
			if !bytes.Equal(weightString, tc.expected) {
				t.Errorf("weight_string(\\uFFFD) = %#v (expected %#v)", weightString, tc.expected)
			}
		})
	}
}

func DebugUcaLegacyWeightString(t *testing.T, collname string, input, expected []byte) {
	type DebugIterator interface {
		uca.WeightIteratorLegacy
		DebugCodepoint() (rune, int)
	}

	coll := testcollation(t, collname).(*Collation_uca_legacy)
	iter := coll.uca.Iterator(input).(DebugIterator)
	defer iter.Done()

	for {
		curcp, _ := iter.DebugCodepoint()
		w, ok := iter.Next()
		if !ok {
			break
		}

		a := byte(w >> 8)
		b := byte(w)
		t.Logf("%q -> %d, %d | %d, %d", string(curcp), a, b, expected[0], expected[1])
		if a != expected[0] || b != expected[1] {
			t.Errorf("mismatch on %q", string(curcp))
		}
		expected = expected[2:]
	}
}

const ExampleString = "abc æøå 日本語"
const ExampleStringLong = "Premature optimization is the root of all evil. " +
	"Våre norske tegn bør æres. 日本語が少しわかります。 " +
	"✌️🐶👩🏽"
const JapaneseString = "データの保存とアクセスを行うストレージエンジンがSQLパーサとは" +
	"分離独立しており、用途に応じたストレージエンジンを選択できる" +
	"「マルチストレージエンジン」方式を採用している。"
const WhitespaceString = "This is a\n prett\ny unrealist\nic case; a\nn " +
	"Eng\nlish sente\nnce where\n we'\nve added a new\nline every te\nn " +
	"bytes or\n so.\n"
const HungarianString = "A MySQL adatbázisok adminisztrációjára a mellékelt " +
	"parancssori eszközöket (mysql és mysqladmin) használhatjuk."
const JapaneseString2 = "サーバー SQL モードの設定方法。この設定は、たとえば" +
	"別のデータベースシステムからのコードとの互換性を保ったり、特定の状況に" +
	"ついてのエラー処理を制御したりするために、SQL の構文およびセマンティクス" +
	"の特定の側面を変更します。"
const ChineseString = "\xE9\x98\xBF\xE5\x92\x97\xF0\xAC\xBA\xA1" +
	"\xC4\x81\x61\x62\xC5\xAB\x75\x55\xC7\x96\x5A\xF0\x94\x99\x86" +
	"\xF0\x97\x86\xA0\xF0\xAC\xBA\xA2\xF0\xAE\xAF\xA0\xF0\xB3\x8C\xB3"
const ChineseString2 = "春江潮水连海平，海上明月共潮生。" +
	"滟滟随波千万里，何处春江无月明！" +
	"江流宛转绕芳甸，月照花林皆似霰；" +
	"空里流霜不觉飞，汀上白沙看不见。" +
	"江天一色无纤尘，皎皎空中孤月轮。" +
	"江畔何人初见月？江月何年初照人？" +
	"人生代代无穷已，江月年年只相似。" +
	"不知江月待何人，但见长江送流水。" +
	"白云一片去悠悠，青枫浦上不胜愁。" +
	"谁家今夜扁舟子？何处相思明月楼？"

var AllTestStrings = []string{
	ExampleString,
	ExampleStringLong,
	JapaneseString,
	WhitespaceString,
	HungarianString,
	JapaneseString2,
	ChineseString,
	ChineseString2,
}

var TestCases = []struct {
	collation string
	input     string
	expected  []byte
}{
	{
		collation: "utf8mb4_bin",
		input:     ExampleString,
		expected: []byte{
			0x00, 0x00, 0x61, 0x00, 0x00, 0x62, 0x00, 0x00, 0x63, // abc
			0x00, 0x00, 0x20, // space
			0x00, 0x00, 0xe6, 0x00, 0x00, 0xf8, 0x00, 0x00, 0xe5, // æøå
			0x00, 0x00, 0x20, // space
			0x00, 0x65, 0xe5, 0x00, 0x67, 0x2c, 0x00, 0x8a, 0x9e, // 日本語
			0x00, 0x00, 0x20, 0x00, 0x00, 0x20, 0x00, 0x00, 0x20,
			0x00, 0x00, 0x20, 0x00, 0x00, 0x20, // space for padding
		},
	},
	{
		collation: "utf8mb4_0900_ai_ci",
		input:     ExampleString,
		expected: []byte{
			0x1c, 0x47, 0x1c, 0x60, 0x1c, 0x7a, // abc
			0x02, 0x09, // space
			0x1c, 0x47, 0x1c, 0xaa, 0x1d, 0xdd, 0x1c, 0x47, // æøå
			0x02, 0x09, // space
			0xfb, 0x40, 0xe5, 0xe5, 0xfb, 0x40, 0xe7, 0x2c,
			0xfb, 0x41, 0x8a, 0x9e, // 日本語
		},
	},
	{
		collation: "utf8mb4_0900_as_ci",
		input:     ExampleString,
		expected: []byte{
			0x1c, 0x47, 0x1c, 0x60, 0x1c, 0x7a, // abc
			0x02, 0x09, // space
			0x1c, 0x47, 0x1c, 0xaa, 0x1d, 0xdd, 0x1c, 0x47, // æøå
			0x02, 0x09, // space
			0xfb, 0x40, 0xe5, 0xe5, 0xfb, 0x40, 0xe7, 0x2c, // 日本語
			0xfb, 0x41, 0x8a, 0x9e, 0x00, 0x00, // level separator
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, // abc
			0x00, 0x20, // space
			0x00, 0x20, 0x01, 0x10, 0x00, 0x20, 0x00, 0x20, // æøå
			0x00, 0x2F, 0x00, 0x20, 0x00, 0x29, 0x00, 0x20, // space
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, // 日本語
		},
	},
	{
		collation: "utf8mb4_0900_as_cs",
		input:     "abc    ",
		expected: []byte{
			0x1c, 0x47, 0x1c, 0x60, 0x1c, 0x7a, // abc
			0x02, 0x09, 0x02, 0x09, 0x02, 0x09,
			0x02, 0x09, // Four spaces.
			0x00, 0x00, // Level separator.
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, // Accents for abc.
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, // Accents for four spaces.
			0x00, 0x00, // Level separator.
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, // Case for abc.
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, // Case for four spaces.
		},
	},
	{
		collation: "utf8mb4_0900_ai_ci",
		input:     ExampleStringLong,
		expected: []byte{
			0x1e, 0x0c, 0x1e, 0x33, 0x1c, 0xaa, 0x1d, 0xaa, 0x1c, 0x47, 0x1e, 0x95,
			0x1e, 0xb5, 0x1e, 0x33, 0x1c, 0xaa, 0x02, 0x09, 0x1d, 0xdd, 0x1e, 0x0c,
			0x1e, 0x95, 0x1d, 0x32, 0x1d, 0xaa, 0x1d, 0x32, 0x1f, 0x21, 0x1c, 0x47,
			0x1e, 0x95, 0x1d, 0x32, 0x1d, 0xdd, 0x1d, 0xb9, 0x02, 0x09, 0x1d, 0x32,
			0x1e, 0x71, 0x02, 0x09, 0x1e, 0x95, 0x1d, 0x18, 0x1c, 0xaa, 0x02, 0x09,
			0x1e, 0x33, 0x1d, 0xdd, 0x1d, 0xdd, 0x1e, 0x95, 0x02, 0x09, 0x1d, 0xdd,
			0x1c, 0xe5, 0x02, 0x09, 0x1c, 0x47, 0x1d, 0x77, 0x1d, 0x77, 0x02, 0x09,
			0x1c, 0xaa, 0x1e, 0xe3, 0x1d, 0x32, 0x1d, 0x77, 0x02, 0x77, 0x02, 0x09,
			0x1e, 0xe3, 0x1c, 0x47, 0x1e, 0x33, 0x1c, 0xaa, 0x02, 0x09, 0x1d, 0xb9,
			0x1d, 0xdd, 0x1e, 0x33, 0x1e, 0x71, 0x1d, 0x65, 0x1c, 0xaa, 0x02, 0x09,
			0x1e, 0x95, 0x1c, 0xaa, 0x1c, 0xf4, 0x1d, 0xb9, 0x02, 0x09, 0x1c, 0x60,
			0x1d, 0xdd, 0x1e, 0x33, 0x02, 0x09, 0x1c, 0x47, 0x1c, 0xaa, 0x1e, 0x33,
			0x1c, 0xaa, 0x1e, 0x71, 0x02, 0x77, 0x02, 0x09, 0xfb, 0x40, 0xe5, 0xe5,
			0xfb, 0x40, 0xe7, 0x2c, 0xfb, 0x41, 0x8a, 0x9e, 0x3d, 0x60, 0xfb, 0x40,
			0xdc, 0x11, 0x3d, 0x66, 0x3d, 0x87, 0x3d, 0x60, 0x3d, 0x83, 0x3d, 0x79,
			0x3d, 0x67, 0x02, 0x8a, 0x02, 0x09, 0x0a, 0x2d, 0x13, 0xdf, 0x14, 0x12,
			0x13, 0xa6,
		},
	},
	{
		collation: "utf8mb4_0900_as_ci",
		input:     ExampleStringLong,
		expected: []byte{
			0x1e, 0x0c, 0x1e, 0x33, 0x1c, 0xaa, 0x1d, 0xaa, 0x1c, 0x47, 0x1e, 0x95,
			0x1e, 0xb5, 0x1e, 0x33, 0x1c, 0xaa, 0x02, 0x09, 0x1d, 0xdd, 0x1e, 0x0c,
			0x1e, 0x95, 0x1d, 0x32, 0x1d, 0xaa, 0x1d, 0x32, 0x1f, 0x21, 0x1c, 0x47,
			0x1e, 0x95, 0x1d, 0x32, 0x1d, 0xdd, 0x1d, 0xb9, 0x02, 0x09, 0x1d, 0x32,
			0x1e, 0x71, 0x02, 0x09, 0x1e, 0x95, 0x1d, 0x18, 0x1c, 0xaa, 0x02, 0x09,
			0x1e, 0x33, 0x1d, 0xdd, 0x1d, 0xdd, 0x1e, 0x95, 0x02, 0x09, 0x1d, 0xdd,
			0x1c, 0xe5, 0x02, 0x09, 0x1c, 0x47, 0x1d, 0x77, 0x1d, 0x77, 0x02, 0x09,
			0x1c, 0xaa, 0x1e, 0xe3, 0x1d, 0x32, 0x1d, 0x77, 0x02, 0x77, 0x02, 0x09,
			0x1e, 0xe3, 0x1c, 0x47, 0x1e, 0x33, 0x1c, 0xaa, 0x02, 0x09, 0x1d, 0xb9,
			0x1d, 0xdd, 0x1e, 0x33, 0x1e, 0x71, 0x1d, 0x65, 0x1c, 0xaa, 0x02, 0x09,
			0x1e, 0x95, 0x1c, 0xaa, 0x1c, 0xf4, 0x1d, 0xb9, 0x02, 0x09, 0x1c, 0x60,
			0x1d, 0xdd, 0x1e, 0x33, 0x02, 0x09, 0x1c, 0x47, 0x1c, 0xaa, 0x1e, 0x33,
			0x1c, 0xaa, 0x1e, 0x71, 0x02, 0x77, 0x02, 0x09, 0xfb, 0x40, 0xe5, 0xe5,
			0xfb, 0x40, 0xe7, 0x2c, 0xfb, 0x41, 0x8a, 0x9e, 0x3d, 0x60, 0xfb, 0x40,
			0xdc, 0x11, 0x3d, 0x66, 0x3d, 0x87, 0x3d, 0x60, 0x3d, 0x83, 0x3d, 0x79,
			0x3d, 0x67, 0x02, 0x8a, 0x02, 0x09, 0x0a, 0x2d, 0x13, 0xdf, 0x14, 0x12,
			0x13, 0xa6, 0x00, 0x00, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x29, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x2F, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x01, 0x10, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x37, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
		},
	},
	{
		collation: "utf8mb4_0900_ai_ci",
		input:     JapaneseString,
		expected: []byte{
			0x3d, 0x6d, 0x1c, 0x0e, 0x3d, 0x6a, 0x3d, 0x73, 0xfb, 0x40, 0xcf, 0xdd,
			0xfb, 0x40, 0xdb, 0x58, 0x3d, 0x6e, 0x3d, 0x5a, 0x3d, 0x62, 0x3d, 0x68,
			0x3d, 0x67, 0x3d, 0x8a, 0xfb, 0x41, 0x88, 0x4c, 0x3d, 0x5c, 0x3d, 0x67,
			0x3d, 0x6e, 0x3d, 0x85, 0x1c, 0x0e, 0x3d, 0x66, 0x3d, 0x5e, 0x3d, 0x8b,
			0x3d, 0x66, 0x3d, 0x8b, 0x3d, 0x60, 0x1e, 0x71, 0x1e, 0x21, 0x1d, 0x77,
			0x3d, 0x74, 0x1c, 0x0e, 0x3d, 0x65, 0x3d, 0x6e, 0x3d, 0x74, 0xfb, 0x40,
			0xd2, 0x06, 0xfb, 0x41, 0x96, 0xe2, 0xfb, 0x40, 0xf2, 0xec, 0xfb, 0x40,
			0xfa, 0xcb, 0x3d, 0x66, 0x3d, 0x6d, 0x3d, 0x5f, 0x3d, 0x83, 0x02, 0x31,
			0xfb, 0x40, 0xf5, 0x28, 0xfb, 0x41, 0x90, 0x14, 0x3d, 0x70, 0xfb, 0x40,
			0xdf, 0xdc, 0x3d, 0x66, 0x3d, 0x6a, 0x3d, 0x67, 0x3d, 0x6e, 0x3d, 0x85,
			0x1c, 0x0e, 0x3d, 0x66, 0x3d, 0x5e, 0x3d, 0x8b, 0x3d, 0x66, 0x3d, 0x8b,
			0x3d, 0x8a, 0xfb, 0x41, 0x90, 0x78, 0xfb, 0x40, 0xe2, 0x9e, 0x3d, 0x6d,
			0x3d, 0x61, 0x3d, 0x84, 0x03, 0x73, 0x3d, 0x79, 0x3d, 0x84, 0x3d, 0x6b,
			0x3d, 0x67, 0x3d, 0x6e, 0x3d, 0x85, 0x1c, 0x0e, 0x3d, 0x66, 0x3d, 0x5e,
			0x3d, 0x8b, 0x3d, 0x66, 0x3d, 0x8b, 0x03, 0x74, 0xfb, 0x40, 0xe5, 0xb9,
			0xfb, 0x40, 0xdf, 0x0f, 0x3d, 0x8a, 0xfb, 0x40, 0xe3, 0xa1, 0xfb, 0x40,
			0xf5, 0x28, 0x3d, 0x66, 0x3d, 0x6d, 0x3d, 0x5b, 0x3d, 0x84, 0x02, 0x8a,
		},
	},
	{
		collation: "utf8mb4_0900_ai_ci",
		input:     WhitespaceString,
		expected: []byte{
			0x1e, 0x95, 0x1d, 0x18, 0x1d, 0x32, 0x1e, 0x71, 0x02, 0x09, 0x1d, 0x32,
			0x1e, 0x71, 0x02, 0x09, 0x1c, 0x47, 0x02, 0x02, 0x02, 0x09, 0x1e, 0x0c,
			0x1e, 0x33, 0x1c, 0xaa, 0x1e, 0x95, 0x1e, 0x95, 0x02, 0x02, 0x1f, 0x0b,
			0x02, 0x09, 0x1e, 0xb5, 0x1d, 0xb9, 0x1e, 0x33, 0x1c, 0xaa, 0x1c, 0x47,
			0x1d, 0x77, 0x1d, 0x32, 0x1e, 0x71, 0x1e, 0x95, 0x02, 0x02, 0x1d, 0x32,
			0x1c, 0x7a, 0x02, 0x09, 0x1c, 0x7a, 0x1c, 0x47, 0x1e, 0x71, 0x1c, 0xaa,
			0x02, 0x34, 0x02, 0x09, 0x1c, 0x47, 0x02, 0x02, 0x1d, 0xb9, 0x02, 0x09,
			0x1c, 0xaa, 0x1d, 0xb9, 0x1c, 0xf4, 0x02, 0x02, 0x1d, 0x77, 0x1d, 0x32,
			0x1e, 0x71, 0x1d, 0x18, 0x02, 0x09, 0x1e, 0x71, 0x1c, 0xaa, 0x1d, 0xb9,
			0x1e, 0x95, 0x1c, 0xaa, 0x02, 0x02, 0x1d, 0xb9, 0x1c, 0x7a, 0x1c, 0xaa,
			0x02, 0x09, 0x1e, 0xf5, 0x1d, 0x18, 0x1c, 0xaa, 0x1e, 0x33, 0x1c, 0xaa,
			0x02, 0x02, 0x02, 0x09, 0x1e, 0xf5, 0x1c, 0xaa, 0x03, 0x05, 0x02, 0x02,
			0x1e, 0xe3, 0x1c, 0xaa, 0x02, 0x09, 0x1c, 0x47, 0x1c, 0x8f, 0x1c, 0x8f,
			0x1c, 0xaa, 0x1c, 0x8f, 0x02, 0x09, 0x1c, 0x47, 0x02, 0x09, 0x1d, 0xb9,
			0x1c, 0xaa, 0x1e, 0xf5, 0x02, 0x02, 0x1d, 0x77, 0x1d, 0x32, 0x1d, 0xb9,
			0x1c, 0xaa, 0x02, 0x09, 0x1c, 0xaa, 0x1e, 0xe3, 0x1c, 0xaa, 0x1e, 0x33,
			0x1f, 0x0b, 0x02, 0x09, 0x1e, 0x95, 0x1c, 0xaa, 0x02, 0x02, 0x1d, 0xb9,
			0x02, 0x09, 0x1c, 0x60, 0x1f, 0x0b, 0x1e, 0x95, 0x1c, 0xaa, 0x1e, 0x71,
			0x02, 0x09, 0x1d, 0xdd, 0x1e, 0x33, 0x02, 0x02, 0x02, 0x09, 0x1e, 0x71,
			0x1d, 0xdd, 0x02, 0x77, 0x02, 0x02,
		},
	},
	{
		collation: "utf8mb4_hu_0900_as_cs",
		input:     HungarianString,
		expected: []byte{
			0x1c, 0x47, 0x02, 0x09, 0x1d, 0xaa, 0x1f, 0x0b, 0x1e, 0x71, 0x1e, 0x21,
			0x1d, 0x77, 0x02, 0x09, 0x1c, 0x47, 0x1c, 0x8f, 0x1c, 0x47, 0x1e, 0x95,
			0x1c, 0x60, 0x1c, 0x47, 0x1f, 0x21, 0x1d, 0x32, 0x1e, 0x71, 0x1d, 0xdd,
			0x1d, 0x65, 0x02, 0x09, 0x1c, 0x47, 0x1c, 0x8f, 0x1d, 0xaa, 0x1d, 0x32,
			0x1d, 0xb9, 0x1d, 0x32, 0x1e, 0x71, 0x54, 0xa5, 0x1e, 0x95, 0x1e, 0x33,
			0x1c, 0x47, 0x1c, 0x7a, 0x1d, 0x32, 0x1d, 0xdd, 0x1d, 0x4c, 0x1c, 0x47,
			0x1e, 0x33, 0x1c, 0x47, 0x02, 0x09, 0x1c, 0x47, 0x02, 0x09, 0x1d, 0xaa,
			0x1c, 0xaa, 0x1d, 0x77, 0x1d, 0x77, 0x1c, 0xaa, 0x1d, 0x65, 0x1c, 0xaa,
			0x1d, 0x77, 0x1e, 0x95, 0x02, 0x09, 0x1e, 0x0c, 0x1c, 0x47, 0x1e, 0x33,
			0x1c, 0x47, 0x1d, 0xb9, 0x1c, 0x7a, 0x54, 0xa5, 0x1e, 0x71, 0x1d, 0xdd,
			0x1e, 0x33, 0x1d, 0x32, 0x02, 0x09, 0x1c, 0xaa, 0x1e, 0x71, 0x54, 0xa5,
			0x1d, 0x65, 0x1d, 0xdd, 0x54, 0xa5, 0x1f, 0x21, 0x1d, 0xdd, 0x54, 0xa5,
			0x1d, 0x65, 0x1c, 0xaa, 0x1e, 0x95, 0x02, 0x09, 0x03, 0x17, 0x1d, 0xaa,
			0x1f, 0x0b, 0x1e, 0x71, 0x1e, 0x21, 0x1d, 0x77, 0x02, 0x09, 0x1c, 0xaa,
			0x1e, 0x71, 0x02, 0x09, 0x1d, 0xaa, 0x1f, 0x0b, 0x1e, 0x71, 0x1e, 0x21,
			0x1d, 0x77, 0x1c, 0x47, 0x1c, 0x8f, 0x1d, 0xaa, 0x1d, 0x32, 0x1d, 0xb9,
			0x03, 0x18, 0x02, 0x09, 0x1d, 0x18, 0x1c, 0x47, 0x1e, 0x71, 0x54, 0xa5,
			0x1d, 0xb9, 0x1c, 0x47, 0x1d, 0x77, 0x1d, 0x18, 0x1c, 0x47, 0x1e, 0x95,
			0x1d, 0x4c, 0x1e, 0xb5, 0x1d, 0x65, 0x02, 0x77, 0x00, 0x00, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x24, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x24, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x24, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x24, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x24, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x24,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x24, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x00, 0x00, 0x08,
			0x00, 0x02, 0x00, 0x08, 0x00, 0x02, 0x00, 0x08, 0x00, 0x08, 0x00, 0x08,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x08, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x08,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x08, 0x00, 0x02, 0x00, 0x08, 0x00, 0x02, 0x00, 0x08, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x08, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
		},
	},
	{
		collation: "utf8mb4_ja_0900_as_cs",
		input:     JapaneseString2,
		expected: []byte{
			0x1F, 0xC1, 0x1F, 0xB6, 0x1F, 0xD0, 0x1F, 0xB6, 0x02, 0x09, 0x1E, 0x71,
			0x1E, 0x21, 0x1D, 0x77, 0x02, 0x09, 0x1F, 0xD9, 0x1F, 0xBB, 0x1F, 0xCA,
			0x1F, 0xCF, 0x5A, 0xC2, 0x5C, 0x45, 0x5E, 0x8C, 0x5E, 0x8E, 0x02, 0x8A,
			0x1F, 0xC0, 0x1F, 0xCF, 0x5A, 0xC2, 0x5C, 0x45, 0x1F, 0xD0, 0x02, 0x31,
			0x1F, 0xC6, 0x1F, 0xCA, 0x1F, 0xBA, 0x1F, 0xD0, 0x5E, 0x5B, 0x1F, 0xCF,
			0x1F, 0xC9, 0x1F, 0xBA, 0x1F, 0xC6, 0x1F, 0xD3, 0x1F, 0xBA, 0x1F, 0xC3,
			0x1F, 0xC2, 0x1F, 0xC3, 0x1F, 0xC9, 0x1F, 0xD7, 0x1F, 0xBC, 0x1F, 0xDE,
			0x1F, 0xCF, 0x1F, 0xC0, 0x1F, 0xBB, 0x1F, 0xCA, 0x1F, 0xCA, 0x1F, 0xCF,
			0x57, 0xD2, 0x56, 0x34, 0x5A, 0x90, 0x1F, 0xE6, 0x5E, 0x6C, 0x1F, 0xC8,
			0x1F, 0xC6, 0x1F, 0xDF, 0x02, 0x31, 0x5C, 0xDA, 0x5C, 0x45, 0x1F, 0xCF,
			0x5A, 0x1C, 0x56, 0xEE, 0x1F, 0xCC, 0x1F, 0xC8, 0x1F, 0xB7, 0x1F, 0xC9,
			0x1F, 0xCF, 0x1F, 0xBA, 0x1F, 0xDE, 0x1F, 0xB6, 0x59, 0xB1, 0x5F, 0xA6,
			0x1F, 0xE6, 0x5A, 0x8C, 0x57, 0xD9, 0x1F, 0xC2, 0x1F, 0xC6, 0x1F, 0xDF,
			0x1F, 0xC3, 0x1F, 0xE0, 0x1F, 0xC6, 0x1F, 0xD8, 0x1F, 0xCC, 0x02, 0x31,
			0x1E, 0x71, 0x1E, 0x21, 0x1D, 0x77, 0x02, 0x09, 0x1F, 0xCF, 0x58, 0x0E,
			0x5E, 0x47, 0x1F, 0xBB, 0x1F, 0xDD, 0x1F, 0xD1, 0x1F, 0xC4, 0x1F, 0xD5,
			0x1F, 0xE7, 0x1F, 0xC9, 0x1F, 0xB7, 0x1F, 0xBE, 0x1F, 0xC3, 0x1F, 0xCF,
			0x5C, 0xDA, 0x5C, 0x45, 0x1F, 0xCF, 0x5B, 0x45, 0x5F, 0x17, 0x1F, 0xE6,
			0x5E, 0x60, 0x58, 0x0A, 0x1F, 0xC2, 0x1F, 0xD5, 0x1F, 0xC3, 0x02, 0x8A,
			0x00, 0x00, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x37, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x37, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x37, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x37,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x37, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x37, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x37,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x00, 0x00, 0x0E, 0x00, 0x0C, 0x00, 0x21,
			0x00, 0x0E, 0x00, 0x02, 0x00, 0x0C, 0x00, 0x21, 0x00, 0x02, 0x00, 0x08,
			0x00, 0x08, 0x00, 0x08, 0x00, 0x02, 0x00, 0x0E, 0x00, 0x0C, 0x00, 0x21,
			0x00, 0x0E, 0x00, 0x02, 0x00, 0x0E, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x0E, 0x00, 0x02, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x0E,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x02, 0x00, 0x0C,
			0x00, 0x21, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x02, 0x00, 0x0C, 0x00, 0x21,
			0x00, 0x0E, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x0E,
			0x00, 0x0E, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x0C, 0x00, 0x21, 0x00, 0x0E,
			0x00, 0x02, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x0E, 0x00, 0x02, 0x00, 0x0D, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x0E, 0x00, 0x02, 0x00, 0x02, 0x00, 0x0E,
			0x00, 0x0E, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x0E,
			0x00, 0x0C, 0x00, 0x21, 0x00, 0x02, 0x00, 0x02, 0x00, 0x0E, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x0E,
			0x00, 0x0E, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x02, 0x00, 0x08, 0x00, 0x08,
			0x00, 0x08, 0x00, 0x02, 0x00, 0x0E, 0x00, 0x02, 0x00, 0x02, 0x00, 0x0E,
			0x00, 0x0E, 0x00, 0x0E, 0x00, 0x02, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x0E,
			0x00, 0x0E, 0x00, 0x0D, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x0E, 0x00, 0x02, 0x00, 0x02, 0x00, 0x0E, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x02,
		},
	},
	{
		collation: "utf8mb4_ja_0900_as_cs_ks",
		input:     JapaneseString2,
		expected: []byte{
			0x1F, 0xC1, 0x1F, 0xB6, 0x1F, 0xD0, 0x1F, 0xB6, 0x02, 0x09, 0x1E, 0x71,
			0x1E, 0x21, 0x1D, 0x77, 0x02, 0x09, 0x1F, 0xD9, 0x1F, 0xBB, 0x1F, 0xCA,
			0x1F, 0xCF, 0x5A, 0xC2, 0x5C, 0x45, 0x5E, 0x8C, 0x5E, 0x8E, 0x02, 0x8A,
			0x1F, 0xC0, 0x1F, 0xCF, 0x5A, 0xC2, 0x5C, 0x45, 0x1F, 0xD0, 0x02, 0x31,
			0x1F, 0xC6, 0x1F, 0xCA, 0x1F, 0xBA, 0x1F, 0xD0, 0x5E, 0x5B, 0x1F, 0xCF,
			0x1F, 0xC9, 0x1F, 0xBA, 0x1F, 0xC6, 0x1F, 0xD3, 0x1F, 0xBA, 0x1F, 0xC3,
			0x1F, 0xC2, 0x1F, 0xC3, 0x1F, 0xC9, 0x1F, 0xD7, 0x1F, 0xBC, 0x1F, 0xDE,
			0x1F, 0xCF, 0x1F, 0xC0, 0x1F, 0xBB, 0x1F, 0xCA, 0x1F, 0xCA, 0x1F, 0xCF,
			0x57, 0xD2, 0x56, 0x34, 0x5A, 0x90, 0x1F, 0xE6, 0x5E, 0x6C, 0x1F, 0xC8,
			0x1F, 0xC6, 0x1F, 0xDF, 0x02, 0x31, 0x5C, 0xDA, 0x5C, 0x45, 0x1F, 0xCF,
			0x5A, 0x1C, 0x56, 0xEE, 0x1F, 0xCC, 0x1F, 0xC8, 0x1F, 0xB7, 0x1F, 0xC9,
			0x1F, 0xCF, 0x1F, 0xBA, 0x1F, 0xDE, 0x1F, 0xB6, 0x59, 0xB1, 0x5F, 0xA6,
			0x1F, 0xE6, 0x5A, 0x8C, 0x57, 0xD9, 0x1F, 0xC2, 0x1F, 0xC6, 0x1F, 0xDF,
			0x1F, 0xC3, 0x1F, 0xE0, 0x1F, 0xC6, 0x1F, 0xD8, 0x1F, 0xCC, 0x02, 0x31,
			0x1E, 0x71, 0x1E, 0x21, 0x1D, 0x77, 0x02, 0x09, 0x1F, 0xCF, 0x58, 0x0E,
			0x5E, 0x47, 0x1F, 0xBB, 0x1F, 0xDD, 0x1F, 0xD1, 0x1F, 0xC4, 0x1F, 0xD5,
			0x1F, 0xE7, 0x1F, 0xC9, 0x1F, 0xB7, 0x1F, 0xBE, 0x1F, 0xC3, 0x1F, 0xCF,
			0x5C, 0xDA, 0x5C, 0x45, 0x1F, 0xCF, 0x5B, 0x45, 0x5F, 0x17, 0x1F, 0xE6,
			0x5E, 0x60, 0x58, 0x0A, 0x1F, 0xC2, 0x1F, 0xD5, 0x1F, 0xC3, 0x02, 0x8A,
			0x00, 0x00, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x37, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x37, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x37, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x37,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x37, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x37, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x37,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x00, 0x00, 0x0E, 0x00, 0x0C, 0x00, 0x21,
			0x00, 0x0E, 0x00, 0x02, 0x00, 0x0C, 0x00, 0x21, 0x00, 0x02, 0x00, 0x08,
			0x00, 0x08, 0x00, 0x08, 0x00, 0x02, 0x00, 0x0E, 0x00, 0x0C, 0x00, 0x21,
			0x00, 0x0E, 0x00, 0x02, 0x00, 0x0E, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x0E, 0x00, 0x02, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x0E,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x02, 0x00, 0x0C,
			0x00, 0x21, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x02, 0x00, 0x0C, 0x00, 0x21,
			0x00, 0x0E, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x0E,
			0x00, 0x0E, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x0C, 0x00, 0x21, 0x00, 0x0E,
			0x00, 0x02, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x0E, 0x00, 0x02, 0x00, 0x0D, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x0E, 0x00, 0x02, 0x00, 0x02, 0x00, 0x0E,
			0x00, 0x0E, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x0E,
			0x00, 0x0C, 0x00, 0x21, 0x00, 0x02, 0x00, 0x02, 0x00, 0x0E, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x0E,
			0x00, 0x0E, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x02, 0x00, 0x08, 0x00, 0x08,
			0x00, 0x08, 0x00, 0x02, 0x00, 0x0E, 0x00, 0x02, 0x00, 0x02, 0x00, 0x0E,
			0x00, 0x0E, 0x00, 0x0E, 0x00, 0x02, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x0E,
			0x00, 0x0E, 0x00, 0x0D, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x0E, 0x00, 0x02, 0x00, 0x02, 0x00, 0x0E, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x0E, 0x00, 0x02, 0x00, 0x00,
			0x00, 0x08, 0x00, 0x08, 0x00, 0x08, 0x00, 0x08, 0x00, 0x08, 0x00, 0x08,
			0x00, 0x08, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x08, 0x00, 0x08,
			0x00, 0x08, 0x00, 0x08, 0x00, 0x08, 0x00, 0x08, 0x00, 0x08, 0x00, 0x08,
			0x00, 0x08, 0x00, 0x08, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x08,
			0x00, 0x08, 0x00, 0x08, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x08, 0x00, 0x08, 0x00, 0x08, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x08, 0x00, 0x08, 0x00, 0x08, 0x00, 0x08, 0x00, 0x08, 0x00, 0x08,
			0x00, 0x08, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02,
		},
	},
	{
		collation: "utf8mb4_zh_0900_as_cs",
		input:     ChineseString,
		expected: []byte{
			0x1C, 0x47, 0xBD, 0xBE, 0xBD, 0xC3, 0xCE, 0xA1, 0xBD, 0xC4, 0xBD, 0xC4,
			0xBD, 0xDD, 0xC0, 0x32, 0xC0, 0x32, 0xC0, 0x32, 0xC0, 0x32, 0xC0, 0x9E,
			0xF6, 0x20, 0xF6, 0x21, 0x81, 0xA0, 0xF6, 0x27, 0xCE, 0xA2, 0xF6, 0x27,
			0xEB, 0xE0, 0xF6, 0x28, 0xB3, 0x33, 0x00, 0x00, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x1F, 0x01, 0x16, 0x00, 0x20, 0x00, 0x20, 0x00, 0x1F,
			0x01, 0x16, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x01, 0x16, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x00,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x08, 0x00, 0x08, 0x00, 0x08, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
		},
	},
	{
		collation: "utf8mb4_zh_0900_as_cs",
		input:     ChineseString2,
		expected: []byte{
			0x2C, 0xD0, 0x4F, 0xF1, 0x28, 0x08, 0x87, 0xE8, 0x60, 0x4C, 0x42, 0xEF,
			0x75, 0x93, 0x02, 0x22, 0x42, 0xEF, 0x83, 0x8A, 0x6C, 0x4F, 0xAF, 0x96,
			0x3F, 0x58, 0x28, 0x08, 0x84, 0xCF, 0x02, 0x8A, 0xA3, 0xA4, 0xA3, 0xA4,
			0x8A, 0x5F, 0x23, 0x71, 0x78, 0xA8, 0x93, 0x1A, 0x5E, 0xD9, 0x02, 0x22,
			0x44, 0xAC, 0x2B, 0xD5, 0x2C, 0xD0, 0x4F, 0xF1, 0x96, 0x31, 0xAF, 0x96,
			0x6C, 0x4F, 0x02, 0x60, 0x4F, 0xF1, 0x63, 0x7B, 0x92, 0xDD, 0xBA, 0x2E,
			0x7F, 0x07, 0x39, 0x15, 0x32, 0xB2, 0x02, 0x22, 0xAF, 0x96, 0xB4, 0x41,
			0x47, 0xD7, 0x62, 0x27, 0x51, 0x4C, 0x85, 0xE9, 0x81, 0x86, 0x02, 0x34,
			0x59, 0x09, 0x5E, 0xD9, 0x63, 0x7B, 0x87, 0xBA, 0x24, 0x78, 0x56, 0x5A,
			0x39, 0x48, 0x02, 0x22, 0x8F, 0x74, 0x83, 0x8A, 0x1E, 0x4D, 0x82, 0x46,
			0x57, 0xD9, 0x24, 0x78, 0x4F, 0x79, 0x02, 0x8A, 0x4F, 0xF1, 0x8E, 0x8A,
			0xA6, 0x3E, 0x81, 0xEE, 0x96, 0x31, 0x99, 0x9E, 0x28, 0x97, 0x02, 0x22,
			0x50, 0xC2, 0x50, 0xC2, 0x59, 0x09, 0xB8, 0x20, 0x3F, 0xCC, 0xAF, 0x96,
			0x66, 0xC9, 0x02, 0x8A, 0x4F, 0xF1, 0x72, 0xB6, 0x44, 0xAC, 0x7F, 0x11,
			0x2B, 0x7B, 0x4F, 0x79, 0xAF, 0x96, 0x02, 0x66, 0x4F, 0xF1, 0xAF, 0x96,
			0x44, 0xAC, 0x6F, 0xD5, 0x2B, 0x7B, 0xB4, 0x41, 0x7F, 0x11, 0x02, 0x66,
			0x7F, 0x11, 0x84, 0xCF, 0x2F, 0xE2, 0x2F, 0xE2, 0x96, 0x31, 0x7B, 0xE1,
			0xA7, 0x41, 0x02, 0x22, 0x4F, 0xF1, 0xAF, 0x96, 0x6F, 0xD5, 0x6F, 0xD5,
			0xB6, 0xC3, 0x9B, 0x15, 0x85, 0xE9, 0x02, 0x8A, 0x24, 0x78, 0xB6, 0x2E,
			0x4F, 0xF1, 0xAF, 0x96, 0x2F, 0xF4, 0x44, 0xAC, 0x7F, 0x11, 0x02, 0x22,
			0x30, 0x86, 0x4F, 0x79, 0xB3, 0xDD, 0x4F, 0xF1, 0x89, 0x2A, 0x63, 0x7B,
			0x87, 0xE8, 0x02, 0x8A, 0x1E, 0x4D, 0xB0, 0x1B, 0xA6, 0x3E, 0x75, 0x00,
			0x7D, 0x93, 0xAB, 0xAF, 0xAB, 0xAF, 0x02, 0x22, 0x7B, 0x7D, 0x3A, 0x63,
			0x76, 0xA2, 0x83, 0x8A, 0x24, 0x78, 0x85, 0x16, 0x2B, 0x2D, 0x02, 0x8A,
			0x84, 0x30, 0x4D, 0xF3, 0x52, 0x63, 0xA5, 0xC7, 0x21, 0xE0, 0xB8, 0x87,
			0xBC, 0x16, 0x02, 0x66, 0x44, 0xAC, 0x2B, 0xD5, 0x9B, 0x15, 0x88, 0x52,
			0x6C, 0x4F, 0xAF, 0x96, 0x64, 0xA1, 0x02, 0x66, 0x00, 0x00, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20,
			0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00, 0x00, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x03,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x03, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x03,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x03, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x03, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x03,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x03, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x03, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x03, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x03,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x03, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x03, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x03,
			0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02, 0x00, 0x02,
			0x00, 0x02, 0x00, 0x03,
		},
	},
}

func TestUCAWeightStrings(t *testing.T) {
	for _, tc := range TestCases {
		t.Run(tc.collation, func(t *testing.T) {
			collation := testcollation(t, tc.collation)
			for maxlen := 0; maxlen < len(tc.expected); maxlen += 2 {
				buf := make([]byte, 0, len(tc.expected))
				result := collation.WeightString(buf, []byte(tc.input), PadToMax)
				if !bytes.Equal(tc.expected[:maxlen], result[:maxlen]) {
					t.Errorf("mismatch at len=%d\ninput:    %#v\nexpected: %#v\nactual:   %#v",
						maxlen, []byte(tc.input), tc.expected[:maxlen], result[:maxlen])
					break
				}
			}
		})
	}
}

func BenchmarkAllUCAWeightStrings(b *testing.B) {
	for _, tc := range TestCases {
		b.Run(tc.collation, func(b *testing.B) {
			collation := testcollation(b, tc.collation)
			buf := make([]byte, 0, len(tc.expected))
			input := []byte(tc.input)

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = collation.WeightString(buf, input, 0)
			}
		})
	}
}

func TestCompareWithWeightString(t *testing.T) {
	var cases = []struct {
		left, right string
		equal       bool
	}{
		{"Abc", "aBC", true},
		{"ǍḄÇ", "ÁḆĈ", false},
		{"\uA73A", "\uA738", false},
		{"\uAC00", "\u326E", true},
	}

	collation := testcollation(t, "utf8mb4_0900_as_ci")

	for _, tc := range cases {
		left := collation.WeightString(nil, []byte(tc.left), 0)
		right := collation.WeightString(nil, []byte(tc.right), 0)

		if bytes.Equal(left, right) != tc.equal {
			t.Errorf("expected %q / %v == %q / %v to be %v", tc.left, left, tc.right, right, tc.equal)
		}
	}
}

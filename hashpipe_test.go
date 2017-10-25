package hashpipe

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testcases = []string{
	"foo",
	"bar",
	"",
	"eyjafjallajökull",
	`Q[F$AAxEdyk<H*|U!Ef}T1Za>_s=fRO~)akWdQb%mu^qxuO06+`,
	`R)cZ;mPYigAS&3H9&;98iv{i/VmFqlv7SZd7=&)>:fqsi[PL^`,
	`cP!ZYOmf5BM<OyiS+@('M4:yNucsu6[Df1?$k=lVd]^N=|RYF`,
	"DonauDampfSchifffahrtsKapitänsMützenStänderLackierVertriebsMitarbeiterVersammlungsOrtsSchildbürgerMeisterPrüfungsLeiterSprosse",
	`:]¦ÐýdÓ(iÕ1y5ÇóCÜ§Y?LÆëñpVfo´±J»SÕ¬]z¡ñàÌêÜØý©jØþ¬Å4J¢áÄ²ñ[ºZã!óAÞòÒÅbÍÙÖïySºÐº(éräjÛ{VtlUßà#rªõÙñ[.NÉçM%P6ä5nÄ¬ð¨UÚXðTÙ°{P>TñÚx2gÞÝx9©ÍãúØz«ûG¦Vú¦¸aÚ±¢ÞÀýá4SôH«i4ô;Èñ¼¦ËïÃÃom³BHQj¢ÅkÝqÒéhmíîÎßðráeT6Ç¨ff;÷Ëh³R&CæzòòI2¾!1òÜeïZ½y6o¼ÞòÒA'ÌMVÜ#0W¦«ºî8E+þ'*½åh³HÑ¡7«L#÷K!ãQÛo®7ºÆ7tO+KLÀFÞxaéxñå¼P.çØ+7"¸»Wù¿þÌËê¾ýþÀ»±¤ì¦ÝFj7^öÊZ,o!19¿¥¤@Ú«¿6¨+ìÊ#È]ÎpaÈBøÄV}Û¹F!¥+?'¸eÀ%J¹ÈN5å[J*:ÔÎLïMt´¿þºIkÍãÖM¾½jâ¨VÃq[ÓknF¥¸OT\ÿFOqW8^µÂ(1öM÷ñHÛB;Zb7yÂDV3½+ßr¨®Ò£¿VTF¹ïk¶YHDóìçÐÈñk¾3dï¸¶!¾¨D¼$a°»döfDõøD~±X\Fè>ð¿åì`,
}

func TestHasHWriter(t *testing.T) {
	hashHelper := func(s string) []byte {
		h := sha256.New()
		h.Write(bytes.NewBufferString(s).Bytes())
		return h.Sum(nil)
	}
	for i, testcase := range testcases {
		t.Run(fmt.Sprintf("Hash and Write %d/%d", i+1, len(testcases)), func(t *testing.T) {
			data := bytes.NewBufferString(testcase)
			wantN := data.Len()
			wantBytes := data.Bytes()
			wantHash := hashHelper(testcase)
			hash := sha256.New()
			resultbuffer := bytes.Buffer{}
			nw := NewWriter(hash)(&resultbuffer)
			n, e := nw.Write(data.Bytes())
			assert.NoError(t, e)
			assert.Equal(t, wantN, n, "Number of bytes that have been written don't match number of bytes in input data, want %d, got %d", wantN, n)
			assert.True(t, bytes.Equal(wantBytes, resultbuffer.Bytes()), "The data that has been written doesn't match the input data, want %s, got %s", wantBytes, resultbuffer.Bytes()) // cannot use assert equal because it fails for []byte{} and []byte(nil) which should be equal
			gotHash := hash.Sum(nil)
			assert.Equal(t, wantHash, gotHash, "The hash of the data doesn't match the hash of the input, want %v, got %v", wantHash, gotHash)
		})
	}

}

func TestHashReader(t *testing.T) {
	hashHelper := func(s string) []byte {
		h := sha256.New()
		h.Write(bytes.NewBufferString(s).Bytes())
		return h.Sum(nil)
	}
	for i, testcase := range testcases {
		t.Run(fmt.Sprintf("Hash and Read %d/%d", i+1, len(testcases)), func(t *testing.T) {
			data := bytes.NewBufferString(testcase)
			wantN := data.Len()
			wantBytes := data.Bytes()
			wantHash := hashHelper(testcase)
			req := httptest.NewRequest("POST", "/", data)
			hash := sha256.New()
			nr := NewReader(hash)(req.Body)
			resultbuffer := make([]byte, data.Len())
			n, e := nr.Read(resultbuffer)
			assert.NoError(t, e)
			assert.Equal(t, wantN, n, "Number of bytes that have been read don't match number of bytes in input data, want %d, got %d", wantN, n)
			assert.Equal(t, wantBytes, resultbuffer, "The data that has been read doesn't match the input data, want %s, got %s", wantBytes, resultbuffer)
			gotHash := hash.Sum(nil)
			assert.Equal(t, wantHash, gotHash, "The hash of the data doesn't match the hash of the input, want %v, got %v", wantHash, gotHash)
		})
	}
}

func BenchmarkHashReader(b *testing.B) {
	sizes := []int{1, 10, 100, 1000, 10000, 100000, 1000000, 10000000}
	for _, size := range sizes {
		b.Run(fmt.Sprintf("Bench: Hashing %d Bytes", size), func(b *testing.B) {
			BenchHashReader(size, b)
		})
	}
}

func BenchHashReader(size int, b *testing.B) {
	data := make([]byte, size)
	output := make([]byte, size)
	rand.Read(data)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hash := sha256.New()
		hpr := NewReader(hash)(bytes.NewBuffer(data))
		for numBytes := 0; numBytes < size; {
			n, e := hpr.Read(output)
			if e != nil && e != io.EOF {
				b.FailNow()
			}
			numBytes += n
		}
	}
}

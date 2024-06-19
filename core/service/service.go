package service

import (
	"capcha/config"
	"capcha/core/domain"
	"capcha/core/ports"
	"capcha/pkg/tr"
	"fmt"
	"github.com/fogleman/gg"
	"github.com/google/uuid"
	"golang.org/x/image/font"
	"image/color"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

type Service struct {
	cnf   *config.ServiceConfig
	cache ports.Cache
}

func New(cnf *config.ServiceConfig, cache ports.Cache) (*Service, error) {
	s := &Service{
		cnf:   cnf,
		cache: cache,
	}
	if err := s.loadFonts(); err != nil {
		return nil, err
	}
	return s, nil
}

var fonts []font.Face

func (s *Service) loadFonts() error {
	d, err := os.ReadDir(s.cnf.Fonts)
	if err != nil {
		return tr.Trace(err)
	}

	for _, file := range d {
		if file.IsDir() {
			continue
		}

		path := filepath.Join(s.cnf.Fonts, file.Name())

		f, err := gg.LoadFontFace(path, float64(s.cnf.FontSize))
		if err != nil {
			return tr.Trace(err)
		}
		fonts = append(fonts, f)

	}

	return nil
}

func (s *Service) CheckCapcha(uid string, checkVal string) (bool, error) {
	val, err := s.cache.Get(uid)

	if err != nil {
		return false, err
	}
	if val == checkVal {
		if err := s.cache.Delete(uid); err != nil {
			return true, err
		}
		if err := s.cache.Delete(uid + "_img"); err != nil {
			return true, err
		}
		return true, nil
	}
	return false, nil
}

func (s *Service) NewCapcha() (*domain.Capcha, error) {
	img, err := s.GenerateCapcha()
	if err != nil {
		return nil, tr.Trace(err)
	}
	uid := uuid.New().String()
	ttl := time.Minute * time.Duration(s.cnf.CapchaTTL)

	capcha := &domain.Capcha{
		Uid:     uid,
		ImgSrc:  fmt.Sprintf("http://127.0.0.1/capcha-image/%s.png", uid),
		Expires: time.Now().Add(ttl).Format(time.RFC3339),
	}

	/*fmt.Println(os.Args[0])
	if err := os.WriteFile("C:\\Users\\79212\\GolandProjects\\capcha\\img\\"+uid+".png", img.Image, 666); err != nil {
		return nil, tr.Trace(err)
	}*/
	if err := s.cache.Set(uid, img.TextVal, ttl); err != nil {
		return nil, tr.Trace(err)
	}
	if err := s.cache.Set(uid+"_img", string(img.Image), ttl); err != nil {
		return nil, tr.Trace(err)
	}
	return capcha, nil
}

func getRandomCharColor() color.Color {
	return color.NRGBA{
		R: randomUint8Range(0, 200),
		G: randomUint8Range(0, 200),
		B: randomUint8Range(0, 200),
		A: randomUint8Range(180, 255),
	}
}

type Capcha struct {
	Image   []byte
	TextVal string
}

func (s *Service) GenerateCapcha() (*Capcha, error) {
	dc := gg.NewContext(s.cnf.CapchaWidth, s.cnf.CapchaHeight)
	dc.SetColor(color.NRGBA{
		R: randomUint8Range(70, 255),
		G: randomUint8Range(70, 255),
		B: randomUint8Range(70, 255),
		A: 255,
	})
	dc.DrawRectangle(0, 0, float64(s.cnf.CapchaWidth), float64(s.cnf.CapchaHeight))
	dc.Fill()

	capcha := &Capcha{}

	var x float64 = 10
	for i := 0; i < s.cnf.LenCapchaText; i++ {
		dc.SetColor(getRandomCharColor())
		dc.SetFontFace(fonts[randomIntRange(0, len(fonts))])
		c := getRandomChar()
		capcha.TextVal += c
		w := drawRotateString(dc, c, x, float64(s.cnf.CapchaHeight+s.cnf.FontSize)/2, float64(randomIntRange(-40, 40)))

		x += w + 8

	}
	s.drawMask(dc)

	dc.Stroke()
	w := &writer{}
	if err := dc.EncodePNG(w); err != nil {
		return nil, err
	}
	capcha.Image = w.Buffer
	return capcha, nil
}

func (s *Service) drawMask(dc *gg.Context) {
	dc.ClearPath()
	dc.SetLineWidth(2)
	for i := 0; i < 5; i++ {
		dc.SetColor(getRandomCharColor())
		dc.DrawLine(
			float64(randomIntRange(0, s.cnf.CapchaWidth)),
			0,
			float64(randomIntRange(0, s.cnf.CapchaWidth)),
			float64(s.cnf.CapchaHeight*2),
		)
		dc.Stroke()
	}
	dc.SetColor(color.NRGBA{
		R: 200,
		G: 200,
		B: 200,
		A: 255,
	})
	dc.SetLineWidth(1)
	for i := 0; i < 70; i++ {
		dc.DrawPoint(
			float64(randomIntRange(0, s.cnf.CapchaWidth)),
			float64(randomIntRange(0, s.cnf.CapchaWidth)),
			1,
		)
	}
	dc.DrawCircle(float64(s.cnf.CapchaWidth), float64(s.cnf.CapchaHeight), 100)
	dc.DrawCircle(float64(s.cnf.CapchaWidth), 0, 100)
	dc.DrawCircle(1, 1, 100)
	dc.Stroke()
}

func drawRotateString(dc *gg.Context, text string, x, y, deg float64) float64 {
	w, h := dc.MeasureString(text)
	dc.RotateAbout(gg.Radians(deg), x+w/2, y-h/2)
	dc.DrawStringAnchored(text, x, y, 0.0, 0.0)
	dc.RotateAbout(gg.Radians(-deg), x+w/2, y+h/2)
	return w
}

type writer struct {
	Buffer []byte
}

func (w *writer) Write(p []byte) (n int, err error) {
	w.Buffer = append(w.Buffer, p...)
	return len(p), nil
}

func randomUint8Range(min, max uint8) uint8 {
	return uint8((rand.Float32() * float32(max-min)) + float32(min))
}

func randomIntRange(max, min int) int {
	return int((rand.Float32() * float32(max-min)) + float32(min))
}

const chars = "ABCDEFGHJKMNPQRSTUVWXYZ123456789"
const lenChars = uint8(len(chars))

func getRandomChar() string {
	return string(chars[randomUint8Range(0, lenChars)])
}

func (s *Service) GetImage(uid string) ([]byte, error) {
	val, err := s.cache.Get(uid + "_img")
	if err != nil {
		return nil, tr.Trace(err)
	}
	return []byte(val), nil
}

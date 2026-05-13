package spatial

import (
	"math"

	"github.com/marcelbudziszewski/paczkomat-predictor/pkg/geo"
)

type Point struct {
	Lat, Lng float64
	ID       int64
	Tag      string
}

type Index struct {
	stepLat, stepLng float64
	minLat, minLng   float64
	maxLat, maxLng   float64
	buckets          map[int64][]Point
	empty            bool
}

func New(pts []Point, bucketM float64) *Index {
	idx := &Index{
		stepLat: bucketM / 111_320.0,
		stepLng: bucketM / 71_500.0,
	}
	if len(pts) == 0 {
		idx.empty = true
		return idx
	}
	idx.minLat = math.Inf(1)
	idx.maxLat = math.Inf(-1)
	idx.minLng = math.Inf(1)
	idx.maxLng = math.Inf(-1)
	for _, p := range pts {
		if p.Lat < idx.minLat {
			idx.minLat = p.Lat
		}
		if p.Lat > idx.maxLat {
			idx.maxLat = p.Lat
		}
		if p.Lng < idx.minLng {
			idx.minLng = p.Lng
		}
		if p.Lng > idx.maxLng {
			idx.maxLng = p.Lng
		}
	}
	idx.buckets = make(map[int64][]Point, len(pts)/2+1)
	for _, p := range pts {
		idx.put(p)
	}
	return idx
}

func (idx *Index) bucketKey(lat, lng float64) int64 {
	ix := int64(math.Floor((lat - idx.minLat) / idx.stepLat))
	iy := int64(math.Floor((lng - idx.minLng) / idx.stepLng))
	return packKey(ix, iy)
}

func packKey(ix, iy int64) int64 {

	return ((ix + 1<<30) << 32) | ((iy + 1<<30) & 0xFFFFFFFF)
}

func (idx *Index) put(p Point) {
	idx.buckets[idx.bucketKey(p.Lat, p.Lng)] = append(idx.buckets[idx.bucketKey(p.Lat, p.Lng)], p)
}

func (idx *Index) Nearest(lat, lng float64) (Point, float64, bool) {
	if idx == nil || idx.empty {
		return Point{}, math.Inf(1), false
	}
	ix := int64(math.Floor((lat - idx.minLat) / idx.stepLat))
	iy := int64(math.Floor((lng - idx.minLng) / idx.stepLng))

	best := math.Inf(1)
	var bestP Point
	found := false

	const maxRings = 64
	for ring := 0; ring <= maxRings; ring++ {
		for dx := -ring; dx <= ring; dx++ {
			for dy := -ring; dy <= ring; dy++ {

				if ring > 0 && abs(dx) != ring && abs(dy) != ring {
					continue
				}
				k := packKey(ix+int64(dx), iy+int64(dy))
				for _, p := range idx.buckets[k] {
					d := geo.Haversine(lat, lng, p.Lat, p.Lng)
					if d < best {
						best = d
						bestP = p
						found = true
					}
				}
			}
		}
		if found {

			ringDist := float64(ring) * math.Max(idx.stepLat*111_320.0, idx.stepLng*71_500.0)
			if best <= ringDist {
				return bestP, best, true
			}
		}
	}
	if found {
		return bestP, best, true
	}
	return Point{}, math.Inf(1), false
}

func (idx *Index) HasWithin(lat, lng float64, radiusM float64) bool {
	if idx == nil || idx.empty {
		return false
	}
	ix := int64(math.Floor((lat - idx.minLat) / idx.stepLat))
	iy := int64(math.Floor((lng - idx.minLng) / idx.stepLng))
	stepM := math.Max(idx.stepLat*111_320.0, idx.stepLng*71_500.0)
	rings := int(math.Ceil(radiusM / stepM))
	for dx := -rings; dx <= rings; dx++ {
		for dy := -rings; dy <= rings; dy++ {
			k := packKey(ix+int64(dx), iy+int64(dy))
			for _, p := range idx.buckets[k] {
				if geo.Haversine(lat, lng, p.Lat, p.Lng) <= radiusM {
					return true
				}
			}
		}
	}
	return false
}

func (idx *Index) Within(lat, lng float64, radiusM float64) []Point {
	if idx == nil || idx.empty {
		return nil
	}
	ix := int64(math.Floor((lat - idx.minLat) / idx.stepLat))
	iy := int64(math.Floor((lng - idx.minLng) / idx.stepLng))

	stepM := math.Max(idx.stepLat*111_320.0, idx.stepLng*71_500.0)
	rings := int(math.Ceil(radiusM / stepM))
	var out []Point
	for dx := -rings; dx <= rings; dx++ {
		for dy := -rings; dy <= rings; dy++ {
			k := packKey(ix+int64(dx), iy+int64(dy))
			for _, p := range idx.buckets[k] {
				if geo.Haversine(lat, lng, p.Lat, p.Lng) <= radiusM {
					out = append(out, p)
				}
			}
		}
	}
	return out
}

func (idx *Index) Len() int {
	if idx == nil || idx.empty {
		return 0
	}
	n := 0
	for _, b := range idx.buckets {
		n += len(b)
	}
	return n
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

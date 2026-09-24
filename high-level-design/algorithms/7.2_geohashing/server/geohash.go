package main

import "strings"

const base32 = "0123456789bcdefghjkmnpqrstuvwxyz"

// Encode converts lat/lng to a geohash string of given precision.
func Encode(lat, lng float64, precision int) string {
	var hash strings.Builder
	minLat, maxLat := -90.0, 90.0
	minLng, maxLng := -180.0, 180.0
	isLng := true
	bit := 0
	ch := 0

	for hash.Len() < precision {
		if isLng {
			mid := (minLng + maxLng) / 2
			if lng >= mid {
				ch = ch*2 + 1
				minLng = mid
			} else {
				ch = ch * 2
				maxLng = mid
			}
		} else {
			mid := (minLat + maxLat) / 2
			if lat >= mid {
				ch = ch*2 + 1
				minLat = mid
			} else {
				ch = ch * 2
				maxLat = mid
			}
		}
		isLng = !isLng
		bit++
		if bit == 5 {
			hash.WriteByte(base32[ch])
			bit = 0
			ch = 0
		}
	}
	return hash.String()
}

// Decode converts a geohash string back to a lat/lng bounding box center.
func Decode(hash string) (lat, lng float64) {
	minLat, maxLat := -90.0, 90.0
	minLng, maxLng := -180.0, 180.0
	isLng := true

	for _, c := range hash {
		idx := strings.IndexByte(base32, byte(c))
		for i := 4; i >= 0; i-- {
			bit := (idx >> i) & 1
			if isLng {
				mid := (minLng + maxLng) / 2
				if bit == 1 {
					minLng = mid
				} else {
					maxLng = mid
				}
			} else {
				mid := (minLat + maxLat) / 2
				if bit == 1 {
					minLat = mid
				} else {
					maxLat = mid
				}
			}
			isLng = !isLng
		}
	}
	return (minLat + maxLat) / 2, (minLng + maxLng) / 2
}

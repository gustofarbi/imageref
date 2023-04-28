package libvips

import (
	"github.com/davidbyttow/govips/v2/vips"
	"gonum.org/v1/gonum/mat"
)

// perspective distortion: each pixel (X, Y) in the output image is
// interpolated from pixel (x, y) in the input using:
//
// X = (x * A + y * B + C) / (x * G + y * H + I)
// Y = (x * D + y * E + F) / (x * G + y * H + I)
//
// where the constants A, B, C.. represent the transform vector, calculated from a set of tie points
func DistortPerspective(ref *vips.ImageRef, tiepoints []float64) error {
	TransformVector, err := calculateTransformation(tiepoints)
	if err != nil {
		return err
	}

	A := TransformVector.AtVec(0)
	B := TransformVector.AtVec(1)
	C := TransformVector.AtVec(2)
	D := TransformVector.AtVec(3)
	E := TransformVector.AtVec(4)
	F := TransformVector.AtVec(5)
	G := TransformVector.AtVec(6)
	H := TransformVector.AtVec(7)
	I := float64(1)

	// index is a two-band image where band 0 stores x coordinate of every pixel and band 1 stores y coordinates
	index, err := vips.XYZ(ref.Width(), ref.Height())
	if err != nil {
		return err
	}

	x, err := extractBand(index, 0, false)
	if err != nil {
		return err
	}

	y, err := extractBand(index, 1, true)
	if err != nil {
		return err
	}

	// x * A
	xA, err := linear(x, A, 0, false)
	if err != nil {
		return err
	}

	// y * B + C
	yB_C, err := linear(y, B, C, false)
	if err != nil {
		return err
	}

	// x * D
	xD, err := linear(x, D, 0, false)
	if err != nil {
		return err
	}

	// y * E + F
	yE_F, err := linear(y, E, F, false)
	if err != nil {
		return err
	}

	// x * A + y * B + C
	xA_yB_C, err := add(xA, yB_C)
	if err != nil {
		return err
	}

	// x * D + y * E + F
	xD_yE_F, err := add(xD, yE_F)
	if err != nil {
		return err
	}

	// x * G
	xG, err := linear(x, G, 0, true)
	if err != nil {
		return err
	}

	// y * H + I
	yH_I, err := linear(y, H, I, true)
	if err != nil {
		return err
	}

	// x * G + y * H + I
	xG_yH_I, err := add(xG, yH_I)
	if err != nil {
		return err
	}

	// X = (x * A + y * B + C) / (x * G + y * H + I)
	X, err := divide(xA_yB_C, xG_yH_I)
	if err != nil {
		return err
	}

	// Y = (x * D + y * E + F) / (x * G + y * H + I)
	Y, err := divide(xD_yE_F, xG_yH_I)
	if err != nil {
		return err
	}

	// join up X and Y back into map image
	mapimage, err := bandjoin(X, Y)
	if err != nil {
		return err
	}

	// transform the original image
	return ref.Mapim(mapimage)
}

func extractBand(src *vips.ImageRef, band int, modifySrc bool) (*vips.ImageRef, error) {
	if modifySrc {
		err := src.ExtractBand(band, 1)
		return src, err
	}
	c, err := src.Copy()
	if err != nil {
		return src, err
	}
	err = c.ExtractBand(band, 1)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func bandjoin(x *vips.ImageRef, y *vips.ImageRef) (*vips.ImageRef, error) {
	err := x.BandJoin(y)
	return x, err
}

func add(left *vips.ImageRef, right *vips.ImageRef) (*vips.ImageRef, error) {
	err := left.Add(right)
	return left, err
}

func linear(image *vips.ImageRef, a float64, b float64, modifySrc bool) (*vips.ImageRef, error) {
	if modifySrc {
		err := image.Linear1(a, b)
		return image, err
	}
	c, err := image.Copy()
	if err != nil {
		return image, err
	}
	err = c.Linear1(a, b)
	return c, err
}

func divide(enumerator, denominator *vips.ImageRef) (*vips.ImageRef, error) {
	c, err := enumerator.Copy()
	if err != nil {
		return c, err
	}
	err = c.Divide(denominator)
	return c, err
}

// Calculates coefficients of perspective transformation
// which maps (xi,yi) to (ui,vi), (i=1,2,3,4):
//
//      c00*xi + c01*yi + c02
// ui = ---------------------
//      c20*xi + c21*yi + c22
//
//      c10*xi + c11*yi + c12
// vi = ---------------------
//      c20*xi + c21*yi + c22
//
// Coefficients are calculated by solving the linear system:
// / x0 y0  1  0  0  0 -x0*u0 -y0*u0 \ /c00\ /u0\
// | x1 y1  1  0  0  0 -x1*u1 -y1*u1 | |c01| |u1|
// | x2 y2  1  0  0  0 -x2*u2 -y2*u2 | |c02| |u2|
// | x3 y3  1  0  0  0 -x3*u3 -y3*u3 |.|c10|=|u3|,
// |  0  0  0 x0 y0  1 -x0*v0 -y0*v0 | |c11| |v0|
// |  0  0  0 x1 y1  1 -x1*v1 -y1*v1 | |c12| |v1|
// |  0  0  0 x2 y2  1 -x2*v2 -y2*v2 | |c20| |v2|
// \  0  0  0 x3 y3  1 -x3*v3 -y3*v3 / \c21/ \v3/
func calculateTransformation(coordinates []float64) (*mat.VecDense, error) {
	// transformation needs to be inversed to work correctly with mapim, therefore the src and target coordinates will be swapped
	// x and y will be assigned target coordinates and u and v will be assigned source coordinates
	u0, v0, x0, y0 := coordinates[0], coordinates[1], coordinates[2], coordinates[3]
	u1, v1, x1, y1 := coordinates[4], coordinates[5], coordinates[6], coordinates[7]
	u2, v2, x2, y2 := coordinates[8], coordinates[9], coordinates[10], coordinates[11]
	u3, v3, x3, y3 := coordinates[12], coordinates[13], coordinates[14], coordinates[15]

	// The data must be arranged in row-major order, i.e. the (i*c + j)-th
	// element in the data slice is the {i, j}-th element in the matrix.
	Adata := []float64{
		x0, y0, 1, 0, 0, 0, -x0 * u0, -y0 * u0,
		x1, y1, 1, 0, 0, 0, -x1 * u1, -y1 * u1,
		x2, y2, 1, 0, 0, 0, -x2 * u2, -y2 * u2,
		x3, y3, 1, 0, 0, 0, -x3 * u3, -y3 * u3,
		0, 0, 0, x0, y0, 1, -x0 * v0, -y0 * v0,
		0, 0, 0, x1, y1, 1, -x1 * v1, -y1 * v1,
		0, 0, 0, x2, y2, 1, -x2 * v2, -y2 * v2,
		0, 0, 0, x3, y3, 1, -x3 * v3, -y3 * v3,
	}
	A := mat.NewDense(8, 8, Adata)
	b := mat.NewVecDense(8, []float64{u0, u1, u2, u3, v0, v1, v2, v3})
	result := &mat.VecDense{}
	err := result.SolveVec(A, b)
	if err != nil {
		return nil, err
	}

	return result, nil
}

package main

// TinyGo version of the WebGL Fundamentals "How it works" lesson
// https://webglfundamentals.org/webgl/lessons/webgl-how-it-works.html

import (
	"math"
	"syscall/js"

	"github.com/justinclift/webgl"
)

var (
	gl *webgl.Context
	width, height int
	program, positionBuffer *js.Value
	positionAttributeLocation int
	matrixLocation *js.Value
	angleInRadians float32
	translateVal, scaleVal []float32

	// Vertex shader source code
	vertCode = `
	attribute vec2 a_position;

	uniform mat3 u_matrix;

	varying vec4 v_color;

	void main() {
		// Multiply the position by the matrix.
		gl_Position = vec4((u_matrix * vec3(a_position, 1)).xy, 0, 1);
		
		// Convert from clipspace to colorspace.
		// Clipspace goes -1.0 to +1.0
		// Colorspace goes from 0.0 to 1.0
		v_color = gl_Position * 0.5 + 0.5;
	}`

	// Fragment shader source code
	fragCode = `
	precision mediump float;
	
	varying vec4 v_color;

	void main() {
		gl_FragColor = v_color;
	}`
)

func main() {
	// Set up the WebGL context
	doc := js.Global().Get("document")
	canvas := doc.Call("getElementById", "mycanvas")
	width = canvas.Get("clientWidth").Int()
	height = canvas.Get("clientHeight").Int()
	canvas.Call("setAttribute", "width", width)
	canvas.Call("setAttribute", "height", height)
	attrs := webgl.DefaultAttributes()
	attrs.Alpha = false
	gl, err := webgl.NewContext(&canvas, attrs)
	if err != nil {
		js.Global().Call("alert", "Error: "+err.Error())
		return
	}

	// * WebGL initialisation code *

	// Create GLSL shaders, upload the GLSL source, compile the shaders
	vertexShader := createShader(gl, webgl.VERTEX_SHADER, vertCode)
	fragmentShader := createShader(gl, webgl.FRAGMENT_SHADER, fragCode)

	// Link the two shaders into a program
	program = createProgram(gl, vertexShader, fragmentShader)

	// Look up where the vertex data needs to go
	positionAttributeLocation = gl.GetAttribLocation(program, "a_position")

	// Look up uniforms
	matrixLocation = gl.GetUniformLocation(program, "u_matrix")

	// Create a buffer
	positionBuffer = gl.CreateArrayBuffer()
	gl.BindBuffer(webgl.ARRAY_BUFFER, positionBuffer)

	// Set Geometry
	setGeometry(gl)

	translateVal = []float32{200, 150}
	angleInRadians = 0
	scaleVal = []float32{1, 1}

	// Setup a ui
	// TODO: Finish converting these pieces to Go
	webglLessonsUI.setupSlider("#x", {value: translateVal[0], slide: updatePosition(0), max: width })
	webglLessonsUI.setupSlider("#y", {value: translateVal[1], slide: updatePosition(1), max: height})
	webglLessonsUI.setupSlider("#angle", {slide: updateAngle, max: 360})
	webglLessonsUI.setupSlider("#scaleX", {value: scale[0], slide: updateScale(0), min: -5, max: 5, step: 0.01, precision: 2})
	webglLessonsUI.setupSlider("#scaleY", {value: scale[1], slide: updateScale(1), min: -5, max: 5, step: 0.01, precision: 2})

}

func updatePosition(i int) {
	return func(event, ui) {
		translation[i] = ui.value
		drawScene()
	}
}

func updateAngle(event, ui) {
	angleInDegrees = 360 - ui.value
	angleInRadians = angleInDegrees * Math.PI / 180
	drawScene()
}

func updateScale(i int) {
	return func(event, ui) {
		scale[i] = ui.value
		drawScene()
	}
}

// WebGL rendering code
func drawScene() {
	// Tell WebGL how to convert from clip space to pixels
	gl.Viewport(0, 0, width, height)

	// Clear the canvas
	// gl.ClearColor(0, 0, 0, 0)
	gl.Clear(webgl.COLOR_BUFFER_BIT)

	// Tell it to use our program (pair of shaders)
	gl.UseProgram(program)

	// Turn on the attribute
	gl.EnableVertexAttribArray(positionAttributeLocation)

	// Bind the position buffer
	gl.BindBuffer(webgl.ARRAY_BUFFER, positionBuffer)

	// Tell the attribute how to get data out of positionBuffer (ARRAY_BUFFER)
	pbSize := 2           // 2 components per iteration
	pbType := webgl.FLOAT // the data is 32bit floats
	pbNormalize := false  // don't normalize the data
	pbStride := 0         // 0 = move forward size * sizeof(pbType) each iteration to get the next position
	pbOffset := 0         // start at the beginning of the buffer
	gl.VertexAttribPointer(positionAttributeLocation, pbSize, pbType, pbNormalize, pbStride, pbOffset)

	// Compute the matrix
	matrix := projection(width, height)
	matrix = translate(matrix, translateVal[0], translateVal[1])
	matrix = rotate(matrix, angleInRadians)
	matrix = scale(matrix, scaleVal[0], scaleVal[1])

	// Set the matrix
	gl.UniformMatrix3fv(matrixLocation, false, matrix)

	// Draw the geometry
	primType := webgl.TRIANGLES
	primOffset := 0
	primCount := 3
	gl.DrawArrays(primType, primOffset, primCount)
}

func createShader(gl *webgl.Context, shaderType int, source string) *js.Value {
	shader := gl.CreateShader(shaderType)
	gl.ShaderSource(shader, source)
	gl.CompileShader(shader)
	success := gl.GetShaderParameter(shader, webgl.COMPILE_STATUS).Bool()
	if success {
		return shader
	}
	println(gl.GetShaderInfoLog(shader))
	gl.DeleteShader(shader)
	return &js.Value{}
}

func createProgram(gl *webgl.Context, vertexShader *js.Value, fragmentShader *js.Value) *js.Value {
	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)
	success := gl.GetProgramParameterb(program, webgl.LINK_STATUS)
	if success {
		return program
	}
	println(gl.GetProgramInfoLog(program))
	gl.DeleteProgram(program)
	return &js.Value{}
}

// Fill the buffer with the values that define a triangle.
// Note, will put the values in whatever buffer is currently
// bound to the ARRAY_BUFFER bind point
func setGeometry(gl *webgl.Context) {
	positionsNative := []float32{
		0, -100,
		150,  125,
		-175,  100,
	}
	positions := js.TypedArrayOf(positionsNative)
	gl.BufferData(webgl.ARRAY_BUFFER, positions, webgl.STATIC_DRAW)
}

/**
 * Takes two Matrix3s, a and b, and computes the product in the order
 * that pre-composes b with a.  In other words, the matrix returned will
 * @param {module:webgl-2d-math.Matrix3} a A matrix.
 * @param {module:webgl-2d-math.Matrix3} b A matrix.
 * @return {module:webgl-2d-math.Matrix3} the result.
 * @memberOf module:webgl-2d-math
 */
func multiply(a, b []float32) []float32 {
	a00 := a[0 * 3 + 0]
	a01 := a[0 * 3 + 1]
	a02 := a[0 * 3 + 2]
	a10 := a[1 * 3 + 0]
	a11 := a[1 * 3 + 1]
	a12 := a[1 * 3 + 2]
	a20 := a[2 * 3 + 0]
	a21 := a[2 * 3 + 1]
	a22 := a[2 * 3 + 2]
	b00 := b[0 * 3 + 0]
	b01 := b[0 * 3 + 1]
	b02 := b[0 * 3 + 2]
	b10 := b[1 * 3 + 0]
	b11 := b[1 * 3 + 1]
	b12 := b[1 * 3 + 2]
	b20 := b[2 * 3 + 0]
	b21 := b[2 * 3 + 1]
	b22 := b[2 * 3 + 2]
	return []float32{
		b00*a00 + b01*a10 + b02*a20,
		b00*a01 + b01*a11 + b02*a21,
		b00*a02 + b01*a12 + b02*a22,
		b10*a00 + b11*a10 + b12*a20,
		b10*a01 + b11*a11 + b12*a21,
		b10*a02 + b11*a12 + b12*a22,
		b20*a00 + b21*a10 + b22*a20,
		b20*a01 + b21*a11 + b22*a21,
		b20*a02 + b21*a12 + b22*a22,
	}
}

/**
 * Creates a 2D projection matrix
 * @param {number} width width in pixels
 * @param {number} height height in pixels
 * @return {[]int} a projection matrix that converts from pixels to clipspace with Y = 0 at the top.
 */
func projection(width, height int) []float32 {
	// Note: This matrix flips the Y axis so 0 is at the top.
	return []float32{
		float32(2 / width), 0, 0,
		0, float32(-2 / height), 0,
		-1, 1, 1,
	}
}

/**
 * Creates a 2D translation matrix
 * @param {number} tx amount to translate in x
 * @param {number} ty amount to translate in y
 * @return {module:webgl-2d-math.Matrix3} a translation matrix that translates by tx and ty.
 * @memberOf module:webgl-2d-math
 */
func translation(tx, ty float32) []float32 {
	return []float32{
		1, 0, 0,
		0, 1, 0,
		tx, ty, 1,
	}
}

/**
 * Multiplies by a 2D translation matrix
 * @param {module:webgl-2d-math.Matrix3} the matrix to be multiplied
 * @param {number} tx amount to translate in x
 * @param {number} ty amount to translate in y
 * @return {module:webgl-2d-math.Matrix3} the result
 * @memberOf module:webgl-2d-math
 */
func translate(m []float32, tx, ty float32) []float32 {
	return multiply(m, translation(tx, ty))
}

/**
 * Creates a 2D rotation matrix
 * @param {number} angleInRadians amount to rotate in radians
 * @return {module:webgl-2d-math.Matrix3} a rotation matrix that rotates by angleInRadians
 * @memberOf module:webgl-2d-math
 */
func rotation(angleInRadians float32) []float32 {
	c := math.Cos(float64(angleInRadians))
	s := math.Sin(float64(angleInRadians))
	return []float32{
		float32(c), float32(-s), 0,
		float32(s), float32(c), 0,
		0, 0, 1,
	}
}

/**
 * Multiplies by a 2D rotation matrix
 * @param {module:webgl-2d-math.Matrix3} the matrix to be multiplied
 * @param {number} angleInRadians amount to rotate in radians
 * @return {module:webgl-2d-math.Matrix3} the result
 * @memberOf module:webgl-2d-math
 */
func rotate(m []float32, angleInRadians float32) []float32 {
	return multiply(m, rotation(angleInRadians))
}

/**
 * Creates a 2D scaling matrix
 * @param {number} sx amount to scale in x
 * @param {number} sy amount to scale in y
 * @return {module:webgl-2d-math.Matrix3} a scale matrix that scales by sx and sy.
 * @memberOf module:webgl-2d-math
 */
func scaling(sx, sy float32) []float32 {
	return []float32{
		sx, 0, 0,
		0, sy, 0,
		0, 0, 1,
	}
}

/**
 * Multiplies by a 2D scaling matrix
 * @param {module:webgl-2d-math.Matrix3} the matrix to be multiplied
 * @param {number} sx amount to scale in x
 * @param {number} sy amount to scale in y
 * @return {module:webgl-2d-math.Matrix3} the result
 * @memberOf module:webgl-2d-math
 */
func scale(m []float32, sx, sy float32) []float32 {
	return multiply(m, scaling(sx, sy))
}

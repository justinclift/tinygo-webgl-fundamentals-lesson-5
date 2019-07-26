package main

// TinyGo version of the WebGL Fundamentals "How it works" lesson
// https://webglfundamentals.org/webgl/lessons/webgl-how-it-works.html

import (
	"math"
	"strconv"
	"syscall/js"

	"github.com/justinclift/webgl"
)

type uiOptions struct {
	Value       float32
	Slide       func(float32, float32)
	Max         float32
	Name        string
	Min         float32
	Step        float32
	Precision   int
	uiPrecision int
	uiMult      float32
}

var (
	gl *webgl.Context
	width, height int
	doc js.Value
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
	doc = js.Global().Get("document")
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
println("0010")
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
	setupSlider("#x", uiOptions{Value: translateVal[0], Slide: updatePosition(0), Max: float32(width) })
	setupSlider("#y", uiOptions{Value: translateVal[1], Slide: updatePosition(1), Max: float32(height)})
	// setupSlider("#angle", uiOptions{Slide: updateAngle, Max: 360})
	setupSlider("#scaleX", uiOptions{Value: scaleVal[0], Slide: updateScale(0), Min: -5, Max: 5, Step: 0.01, Precision: 2})
	setupSlider("#scaleY", uiOptions{Value: scaleVal[1], Slide: updateScale(1), Min: -5, Max: 5, Step: 0.01, Precision: 2})
	println("0020")
	gl.Viewport(0, 0, width, height)
	println("0025")
	// drawScene()

	gl.Clear(webgl.COLOR_BUFFER_BIT)
	println("3020")
	// Tell it to use our program (pair of shaders)
	gl.UseProgram(program)
	println("3030")
	// Turn on the attribute
	gl.EnableVertexAttribArray(positionAttributeLocation)
	println("3040")
	// Bind the position buffer
	gl.BindBuffer(webgl.ARRAY_BUFFER, positionBuffer)
	println("3050")
	// Tell the attribute how to get data out of positionBuffer (ARRAY_BUFFER)
	pbSize := 2           // 2 components per iteration
	pbType := webgl.FLOAT // the data is 32bit floats
	pbNormalize := false  // don't normalize the data
	pbStride := 0         // 0 = move forward size * sizeof(pbType) each iteration to get the next position
	pbOffset := 0         // start at the beginning of the buffer
	println("3060")
	gl.VertexAttribPointer(positionAttributeLocation, pbSize, pbType, pbNormalize, pbStride, pbOffset)
	println("3070")
	// Compute the matrix
	matrix := projection(width, height)
	println("3080")
	matrix = translate(matrix, translateVal[0], translateVal[1])
	println("3090")
	matrix = rotate(matrix, angleInRadians)
	println("3100")
	matrix = scale(matrix, scaleVal[0], scaleVal[1])
	println("3110")
	// Set the matrix
	gl.UniformMatrix3fv(matrixLocation, false, matrix)
	println("3120")
	// Draw the geometry
	primType := webgl.TRIANGLES
	primOffset := 0
	primCount := 3
	println("3130")
	gl.DrawArrays(primType, primOffset, primCount)


	println("0030")
}

//go:export updatePosition
func updatePosition(i int) func(event, ui float32) {
	return func(event, ui float32) {
		translateVal[i] = ui
		drawScene()
	}
}

//go:export updateAngle
func updateAngle(event, ui float32) {
	angleInDegrees := 360 - ui
	angleInRadians = angleInDegrees * math.Pi / 180
	drawScene()
}

//go:export updateScale
func updateScale(i int) func(event, ui float32) {
	return func(event, ui float32) {
		scaleVal[i] = ui
		drawScene()
	}
}

// WebGL rendering code
func drawScene() {
	println("3000")
	// Tell WebGL how to convert from clip space to pixels
	println("Width: " + strconv.FormatInt(int64(width), 10))
	println("Height: " + strconv.FormatInt(int64(height), 10))
	gl.Viewport(0, 0, width, height)
	println("3010")
	// Clear the canvas
	// gl.ClearColor(0, 0, 0, 0)
	gl.Clear(webgl.COLOR_BUFFER_BIT)
	println("3020")
	// Tell it to use our program (pair of shaders)
	gl.UseProgram(program)
	println("3030")
	// Turn on the attribute
	gl.EnableVertexAttribArray(positionAttributeLocation)
	println("3040")
	// Bind the position buffer
	gl.BindBuffer(webgl.ARRAY_BUFFER, positionBuffer)
	println("3050")
	// Tell the attribute how to get data out of positionBuffer (ARRAY_BUFFER)
	pbSize := 2           // 2 components per iteration
	pbType := webgl.FLOAT // the data is 32bit floats
	pbNormalize := false  // don't normalize the data
	pbStride := 0         // 0 = move forward size * sizeof(pbType) each iteration to get the next position
	pbOffset := 0         // start at the beginning of the buffer
	println("3060")
	gl.VertexAttribPointer(positionAttributeLocation, pbSize, pbType, pbNormalize, pbStride, pbOffset)
	println("3070")
	// Compute the matrix
	matrix := projection(width, height)
	println("3080")
	matrix = translate(matrix, translateVal[0], translateVal[1])
	println("3090")
	matrix = rotate(matrix, angleInRadians)
	println("3100")
	matrix = scale(matrix, scaleVal[0], scaleVal[1])
	println("3110")
	// Set the matrix
	gl.UniformMatrix3fv(matrixLocation, false, matrix)
	println("3120")
	// Draw the geometry
	primType := webgl.TRIANGLES
	primOffset := 0
	primCount := 3
	println("3130")
	gl.DrawArrays(primType, primOffset, primCount)
	println("3140")
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

func setupSlider(selector string, options uiOptions) {
println("1000")
	var parent = doc.Call("querySelector", selector)
println("1010")
	if parent == js.Undefined() {
		println("1020")
		return // like jquery don't fail on a bad selector
	}
	println("1030")
	if options.Name == "" {
		println("1040")
		options.Name = selector[1:]
		// options.Name = selector.substring(1)
	}
	println("1050")
	createSlider(parent, options)
	println("1060")
	return
}

func updateValue(e js.Value, value float32, step float32, mult float32, precision int) {
	// Calculate the raw value for the new value
	newVal := float64(value * step * mult)
	e.Set("textContent", strconv.FormatFloat(newVal, 'f', precision, 32))
}

//go:export handleChange
func handleChange(value int) {
// func handleChange(event js.Value) {
	println("handleChange called with value of: " + strconv.FormatInt(int64(value), 10))
	// var value = parseInt(event.target.value)
	// updateValue(value);
	// fn(event, { value: value * step });
}

func createSlider(parent js.Value, options uiOptions) {
	println("2000")
	var step float32
	var max float32
	var uiPrecision int
	var uiMult float32
	println("2010")
	precision := options.Precision
	println("2020")
	min := options.Min
	println("2030")
	value := options.Value
	println("2040")
	// fn := options.Slide
	name := options.Name
	// gopt == getQueryParams().  Figure it out later
	// name := gopt["ui-" + options.name] || options.Name
	if options.Step != 0 {
		println("2050")
		step = 1
	}
	println("2060")
	if options.Max == 0 {
		println("2070")
		max = 1
	}
	println("2080")
	// Not an exact equivalent for the JS ternary, but should be ok to start with
	if options.uiPrecision == 0 {
		println("2090")
		uiPrecision = precision
		println("2100")
	} else {
		println("2110")
		uiPrecision = options.uiPrecision
		println("2120")
	}
	println("2130")
	if options.uiMult == 0 {
		println("2140")
		uiMult = 1
		println("2150")
	}
	println("2160")
	min /= step
	println("2170")
	max /= step
	println("2180")
	value /= step
	println("2190")

	parent.Set("innerHTML", `<div class="gman-widget-outer">
		<div class="gman-widget-label">` + name + `</div>
		<div class="gman-widget-value"></div>
		<input class="gman-widget-slider" type="range" min="` + strconv.FormatFloat(float64(min), 'f', 1, 32) + `" max="` + strconv.FormatFloat(float64(max), 'f', 1, 32) + `" value="` + strconv.FormatFloat(float64(value), 'f', 1, 32) + `" />
	</div>`)
	println("2200")
	// parent.innerHTML = `
    //   <div class="gman-widget-outer">
    //     <div class="gman-widget-label">${name}</div>
    //     <div class="gman-widget-value"></div>
    //     <input class="gman-widget-slider" type="range" min="${min}" max="${max}" value="${value}" />
    //   </div>`;
	valueElem := parent.Call("querySelector", ".gman-widget-value")
	// var valueElem = parent.querySelector(".gman-widget-value");
	println("2210")
	sliderElem := parent.Call("querySelector", ".gman-widget-slider")
	// var sliderElem = parent.querySelector(".gman-widget-slider");
	println("2220")
	updateValue(valueElem, value, step, uiMult, uiPrecision)
	println("2230")
	hcJS := js.Global().Get("handleChange")
	sliderElem.Call("addEventListener", "input", hcJS)
	// sliderElem.addEventListener('input', handleChange);
	println("2240")
	sliderElem.Call("addEventListener", "change", hcJS)
	// sliderElem.addEventListener('change', handleChange);
	println("2250")
	// return {
	// elem: parent,
	// 	updateValue: (v) => {
	// 		v /= step;
	// 		sliderElem.value = v;
	// 		updateValue(v);
	// 	},
	// };
}

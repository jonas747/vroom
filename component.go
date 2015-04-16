package vroom

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/vova616/chipmunk"
	"github.com/vova616/chipmunk/vect"
)

type Component interface {
	Name() string

	AddComponent(component Component)
	RemoveComponent(component Component)

	GetComponents() map[string][]Component
	GetComponentsByName(name string) []Component
	GetComponent(name string) Component

	GetParent() Entity
	SetParent(ent Entity)

	InitCalled() bool
	SetInitCalled()
	Init()

	SetEnabled(bool)
	Enabled() bool
	Destroy()
}

type BaseComponent struct {
	Components map[string][]Component
	IsDisabled bool
	Parent     Entity
	IsInit     bool
}

func (bc *BaseComponent) InitCalled() bool {
	return bc.IsInit
}

func (bc *BaseComponent) Init() {}

func (bc *BaseComponent) SetInitCalled() {
	bc.IsInit = true
}

func (bc *BaseComponent) AddComponent(component Component) {
	bc.Parent.AddComponent(component)
}

func (bc *BaseComponent) RemoveComponent(component Component) {
	bc.Parent.RemoveComponent(component)
}

func (bc *BaseComponent) GetComponents() map[string][]Component {
	return bc.Parent.GetComponents()
}

func (bc *BaseComponent) GetComponentsByName(name string) []Component {
	return bc.Parent.GetComponentsByName(name)
}

func (bc *BaseComponent) GetComponent(name string) Component {
	return bc.Parent.GetComponent(name)
}

func (bc *BaseComponent) GetParent() Entity {
	return bc.Parent
}

func (bc *BaseComponent) SetEnabled(enabled bool) {
	bc.IsDisabled = !enabled
}

func (bc *BaseComponent) Enabled() bool {
	return !bc.IsDisabled
}

func (bc *BaseComponent) SetParent(ent Entity) {
	bc.Parent = ent
}

func (bc *BaseComponent) Destroy() {

}

// Core components
type Transform struct {
	BaseComponent
	Position vect.Vect
	Angle    vect.Float
	Scale    float32
}

func NewTransform(x, y, angle vect.Float) *Transform {
	return &Transform{
		Position: vect.Vect{x, y},
		Angle:    angle,
	}
}

func (t *Transform) Name() string {
	return "Transform"
}

func (t *Transform) CalcPos() vect.Vect {
	physComp := t.GetComponent("PhysBodyComp")
	if physComp != nil {
		casted := physComp.(*PhysBodyComp)
		if casted.Body != nil {
			pos := casted.Body.Position()
			return pos
		}
	}

	parentEntity := t.GetParent().GetParent()
	if parentEntity != nil {
		transformComp := parentEntity.GetComponent("Transform")
		if transformComp != nil {
			casted := transformComp.(*Transform)
			copy := t.Position
			copy.Add(casted.CalcPos())
			return copy
		}
	}

	return t.Position
}

func (t *Transform) GetScreenPos() vect.Vect {
	realPos := t.CalcPos()
	realPos.Add(t.GetParent().GetEngine().Camera)
	return realPos
}

func (t *Transform) CalcAngle() vect.Float {
	physComp := t.GetComponent("PhysBodyComp")
	if physComp != nil {
		casted := physComp.(*PhysBodyComp)
		if casted.Body != nil {
			angle := casted.Body.Angle() * chipmunk.DegreeConst
			return angle
		}
	}

	parentEntity := t.GetParent().GetParent()
	if parentEntity != nil {

		transformComp := parentEntity.GetComponent("Transform")
		if transformComp != nil {
			casted := transformComp.(*Transform)
			copy := t.Angle
			copy += casted.CalcAngle()
			return copy
		}
	}
	return t.Angle
}

type UpdateAble interface {
	Component
	Update(dt float64)
}

// So you can add callbacks direcly to the entity (dont do this)
type UpdateComp struct {
	Component
	OnUpdate func(dt float64)
}

func (upd *UpdateComp) Name() string {
	return "UpdateComp"
}

func (upd *UpdateComp) update(dt float64) {
	if upd.OnUpdate != nil {
		upd.OnUpdate(dt)
	}
}

type DrawAble interface {
	Component
	Draw(renderer *sdl.Renderer)
	GetLayer() int
}

// So you can add callbacks direcly to the entity (dont do this)
type DrawComp struct {
	BaseComponent
	OnDraw func(renderer *sdl.Renderer)
	Layer  int
}

func (drw *DrawComp) Name() string {
	return "DrawComp"
}

func (drw *DrawComp) Draw(renderer *sdl.Renderer) {
	if drw.OnDraw != nil {
		drw.OnDraw(renderer)
	}
}

func (drw *DrawComp) GetLayer() int {
	return drw.GetLayer()
}

type MouseBox struct { // If the mouse is inside this events will be sent
	BaseComponent
	Active bool
	W      int
	H      int
}

func (mb *MouseBox) Name() string {
	return "MouseBox"
}

// Mouse click listener
// If entity also has mbox component will only send clicks inside said mbox
type MouseClickListener interface {
	Component
	MouseDown(x, y, button int)
	MouseUp(x, y, button int)
}

// Move is called when moving isnide mbox
type MouseHoverListener interface {
	Component
	MouseMove(x, y int)
	MouseEnter() // Called whenever the mouse enters the mbox
	MouseLeave() // When leaving
}

type KeyboardListener interface {
	Component
	KeyDown(sdl.Keycode)
	KeyUp(sdl.Keycode)
}

type KeyboardComp struct {
	BaseComponent
	OnKeyDown func(sdl.Keycode)
	OnKeyUp   func(sdl.Keycode)
}

func (kb *KeyboardComp) KeyDown(code sdl.Keycode) {
	if kb.OnKeyDown != nil {
		kb.OnKeyDown(code)
	}
}

func (kb *KeyboardComp) KeyUp(code sdl.Keycode) {
	if kb.OnKeyUp != nil {
		kb.OnKeyUp(code)
	}
}

func (kb *KeyboardComp) Name() string {
	return "KeyboardComp"
}

type PhysBodyComp struct {
	BaseComponent
	Body                 *chipmunk.Body
	CollisionEnterCB     func(arbiter *chipmunk.Arbiter) bool
	CollisionPreSolveCB  func(arbiter *chipmunk.Arbiter) bool
	CollisionPostSolveCB func(arbiter *chipmunk.Arbiter)
	CollisionExitCB      func(arbiter *chipmunk.Arbiter)
}

func (pb *PhysBodyComp) Name() string {
	return "PhysBodyComp"
}

func (pb *PhysBodyComp) CreateBoxBody(w, h int, mass, moment vect.Float, static bool) {
	if pb.Body != nil {
		pb.GetParent().GetEngine().Space.RemoveBody(pb.Body)
	}

	transform := pb.GetComponent("Transform").(*Transform)
	pos := transform.CalcPos()
	angle := transform.CalcAngle()

	shape := chipmunk.NewBox(vect.Vect{vect.Float(w) / 2, vect.Float(h) / 2.0}, vect.Float(w), vect.Float(h))

	var body *chipmunk.Body
	if static {
		body = chipmunk.NewBodyStatic()
	} else {
		body = chipmunk.NewBody(mass, moment)
	}

	body.AddShape(shape)
	body.SetAngle(angle * chipmunk.RadianConst)
	body.SetPosition(pos)

	pb.Body = body
	body.CallbackHandler = pb
}

func (pb *PhysBodyComp) CreateCircleBody(w, h int, static bool) {

}

func (pb *PhysBodyComp) CollisionEnter(arbiter *chipmunk.Arbiter) bool {
	if pb.CollisionEnterCB != nil {
		return pb.CollisionEnterCB(arbiter)
	}
	return true
}
func (pb *PhysBodyComp) CollisionPreSolve(arbiter *chipmunk.Arbiter) bool {
	if pb.CollisionPreSolveCB != nil {
		return pb.CollisionPreSolveCB(arbiter)
	}
	return true
}
func (pb *PhysBodyComp) CollisionPostSolve(arbiter *chipmunk.Arbiter) {
	if pb.CollisionPostSolveCB != nil {
		pb.CollisionPostSolveCB(arbiter)
	}
}
func (pb *PhysBodyComp) CollisionExit(arbiter *chipmunk.Arbiter) {
	if pb.CollisionExitCB != nil {
		pb.CollisionExitCB(arbiter)
	}
}
func (pb *PhysBodyComp) Destroy() {
	if pb.Body != nil {
		pb.GetParent().GetEngine().Space.RemoveBody(pb.Body)
	}
	pb.BaseComponent.Destroy()
}

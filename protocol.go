package openpoker

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"reflect"
)

const (
	GOOD               Cmd = 0
	LOGIN              Cmd = 1
	LOGOUT             Cmd = 2
	BAD                Cmd = 255
	START_GAME         Cmd = 17
	YOU_ARE            Cmd = 31
	YOUR_GAME          Cmd = 39
	SEAT_QUERY         Cmd = 14
	SEAT_INFO          Cmd = 30
	GAME_QUERY         Cmd = 13
	GAME_INFO          Cmd = 18
	JOIN               Cmd = 8
	NOTIFY_JOIN        Cmd = 44
	WATCH              Cmd = 3
	NOTIFY_CANCEL_GAME Cmd = 25
	UNWATCH            Cmd = 4
	LEAVE              Cmd = 9
	NOTIFY_LEAVE       Cmd = 45
	NOTIFY_START_GAME  Cmd = 23
	NOTIFY_BUTTON      Cmd = 35
	NOTIFY_SB          Cmd = 36
	BET_REQ            Cmd = 20
	RAISE              Cmd = 6
	BALANCE_QUERY      Cmd = 16
	BALANCE            Cmd = 33
	FOLD               Cmd = 7
	NOTIFY_RAISE       Cmd = 42
	NOTIFY_BB          Cmd = 37
	NOTIFY_DRAW        Cmd = 21
	NOTIFY_SHARED      Cmd = 22
	NOTIFY_HAND        Cmd = 27
	NOTIFY_END_GAME    Cmd = 24
	SIT_OUT            Cmd = 10
	COME_BACK          Cmd = 11
	CHAT               Cmd = 12
	NOTIFY_CHAT        Cmd = 43
	GAME_STAGE         Cmd = 29
	SHOW_CARDS         Cmd = 40
	NOTIFY_WIN         Cmd = 26
	PLAYER_QUERY       Cmd = 15
	PLAYER_INFO        Cmd = 19

	GT_TEXAS_HOLDEM GameType = 1

	LIMIT_FIXED LimitType = 1
	LIMIT_NONE  LimitType = 2
	LIMIT_POT   LimitType = 3

	OP_IGNORE Op = 0
	OP_EQ     Op = 1
	OP_LT     Op = 2
	OP_GT     Op = 3

	GS_PREFLOP       GameStage = 0
	GS_FLOP          GameStage = 1
	GS_TURN          GameStage = 2
	GS_RIVER         GameStage = 3
	GS_DELAYED_START GameStage = 4
	GS_BLINDS        GameStage = 5
	GS_SHOWDOWN      GameStage = 6

	PS_EMPTY     PlayerState = 0
	PS_PLAY      PlayerState = 1
	PS_FOLD      PlayerState = 2
	PS_WAIT_BB   PlayerState = 4
	PS_SIT_OUT   PlayerState = 8
	PS_MAKEUP_BB PlayerState = 16
	PS_ALL_IN    PlayerState = 32
	PS_BET       PlayerState = 64
	PS_RESERVED  PlayerState = 128 // reserved seat
	// these can never be sent since playerstate is passed as a single byte?!
	// PS_AUTOPLAY  PlayerState = 256
	// PS_MUCK      PlayerState = 512  // will show cards
	// PS_OUT       PlayerState = 1024 // can't play anymore

	SQ_2  Seq = 1
	SQ_3  Seq = 2
	SQ_4  Seq = 3
	SQ_5  Seq = 4
	SQ_6  Seq = 5
	SQ_7  Seq = 6
	SQ_8  Seq = 7
	SQ_9  Seq = 8
	SQ_10 Seq = 9
	SQ_J  Seq = 10
	SQ_Q  Seq = 11
	SQ_K  Seq = 12
	SQ_A  Seq = 13

	CLUBS    Suit = 1
	DIAMONDS Suit = 2
	HEARTS   Suit = 3
	SPADES   Suit = 4
)

var msgTypes = map[Cmd]reflect.Type{
	GOOD:               reflect.TypeOf((*MsgGood)(nil)).Elem(),
	BAD:                reflect.TypeOf((*MsgBad)(nil)).Elem(),
	YOU_ARE:            reflect.TypeOf((*MsgYouAre)(nil)).Elem(),
	YOUR_GAME:          reflect.TypeOf((*MsgYourGame)(nil)).Elem(),
	SEAT_INFO:          reflect.TypeOf((*MsgSeatInfo)(nil)).Elem(),
	GAME_INFO:          reflect.TypeOf((*MsgGameInfo)(nil)).Elem(),
	NOTIFY_JOIN:        reflect.TypeOf((*MsgNotifyJoin)(nil)).Elem(),
	NOTIFY_CANCEL_GAME: reflect.TypeOf((*MsgNotifyCancelGame)(nil)).Elem(),
	NOTIFY_LEAVE:       reflect.TypeOf((*MsgNotifyLeave)(nil)).Elem(),
	NOTIFY_START_GAME:  reflect.TypeOf((*MsgNotifyStartGame)(nil)).Elem(),
	NOTIFY_BUTTON:      reflect.TypeOf((*MsgNotifyButton)(nil)).Elem(),
	NOTIFY_SB:          reflect.TypeOf((*MsgNotifySb)(nil)).Elem(),
	BET_REQ:            reflect.TypeOf((*MsgBetReq)(nil)).Elem(),
	BALANCE:            reflect.TypeOf((*MsgBalance)(nil)).Elem(),
	NOTIFY_RAISE:       reflect.TypeOf((*MsgNotifyRaise)(nil)).Elem(),
	NOTIFY_BB:          reflect.TypeOf((*MsgNotifyBb)(nil)).Elem(),
	NOTIFY_DRAW:        reflect.TypeOf((*MsgNotifyDraw)(nil)).Elem(),
	NOTIFY_SHARED:      reflect.TypeOf((*MsgNotifyShared)(nil)).Elem(),
	NOTIFY_HAND:        reflect.TypeOf((*MsgNotifyHand)(nil)).Elem(),
	NOTIFY_END_GAME:    reflect.TypeOf((*MsgNotifyEndGame)(nil)).Elem(),
	NOTIFY_CHAT:        reflect.TypeOf((*MsgNotifyChat)(nil)).Elem(),
	GAME_STAGE:         reflect.TypeOf((*MsgGameStage)(nil)).Elem(),
	SHOW_CARDS:         reflect.TypeOf((*MsgShowCards)(nil)).Elem(),
	NOTIFY_WIN:         reflect.TypeOf((*MsgNotifyWin)(nil)).Elem(),
}

type Cmd uint8
type Text string
type Small uint8
type Big uint32
type Amount float32
type Op byte
type LimitType byte
type PlayerState byte
type GameStage byte
type GameType byte
type Id int32
type Seq byte
type Suit byte
type Card struct {
	Seq  Seq
	Suit Suit
}
type Cards []Card

type writable interface {
	write(w io.Writer)
}

type readable interface {
	read(r io.Reader) uint16
}

func (l *Cards) read(r io.Reader) uint16 {
	n := readByte(r)
	*l = make([]Card, n)
	v := *l
	for i := uint8(0); i < n; i++ {
		v[i] = Card{
			Seq(readByte(r)),
			Suit(readByte(r)),
		}
	}

	return (2 * uint16(n)) + 1
}

func (c *Cmd) read(r io.Reader) uint16 {
	*c = Cmd(readByte(r))
	return 1
}

func (c *Text) read(r io.Reader) uint16 {
	s, l := readString(r)
	*c = Text(s)
	return l
}

func (c *Small) read(r io.Reader) uint16 {
	*c = Small(readByte(r))
	return 1
}

func (b *Big) read(r io.Reader) uint16 {
	*b = Big(readInt(r))
	return 4
}

func (a *Amount) read(r io.Reader) uint16 {
	*a = Amount(readInt(r)) / 100
	return 4
}

func (c *Op) read(r io.Reader) uint16 {
	*c = Op(readByte(r))
	return 1
}

func (c *LimitType) read(r io.Reader) uint16 {
	*c = LimitType(readByte(r))
	return 1
}

func (c *GameStage) read(r io.Reader) uint16 {
	*c = GameStage(readByte(r))
	return 1
}

func (c *PlayerState) read(r io.Reader) uint16 {
	*c = PlayerState(readByte(r))
	return 1
}

func (c *GameType) read(r io.Reader) uint16 {
	*c = GameType(readByte(r))
	return 1
}

func (c *Seq) read(r io.Reader) uint16 {
	*c = Seq(readByte(r))
	return 1
}

func (c *Suit) read(r io.Reader) uint16 {
	*c = Suit(readByte(r))
	return 1
}

func (c *Id) read(r io.Reader) uint16 {
	*c = Id(readInt(r))
	return 4
}

func (o LimitType) write(w io.Writer) {
	binary.Write(w, binary.BigEndian, o)
}

func (i Id) write(w io.Writer) {
	binary.Write(w, binary.BigEndian, i)
}

func (g GameType) write(w io.Writer) {
	binary.Write(w, binary.BigEndian, g)
}

func (o Op) write(w io.Writer) {
	binary.Write(w, binary.BigEndian, o)
}

func (c Cmd) write(w io.Writer) {
	binary.Write(w, binary.BigEndian, c)
}

func (n Small) write(w io.Writer) {
	binary.Write(w, binary.BigEndian, n)
}

func (n Big) write(w io.Writer) {
	binary.Write(w, binary.BigEndian, n)
}

func (n Amount) write(w io.Writer) {
	Big(n * 100).write(w)
}

func (t Text) write(w io.Writer) {
	var bytes = []byte(t)
	binary.Write(w, binary.BigEndian, uint8(len(bytes)))
	binary.Write(w, binary.BigEndian, bytes)
}

func readByte(r io.Reader) uint8 {
	var data uint8
	binary.Read(r, binary.BigEndian, &data)
	return data
}

func readInt(r io.Reader) uint32 {
	var data uint32
	binary.Read(r, binary.BigEndian, &data)
	return data
}

func readString(r io.Reader) (string, uint16) {
	var size uint8
	binary.Read(r, binary.BigEndian, &size)
	var bytes = make([]byte, size)
	io.ReadFull(r, bytes)
	return string(bytes), uint16(size + 1)
}

func writeMessage(w io.Writer, body ...writable) {

	// fmt.Printf("-> %v\n", body)

	var buf bytes.Buffer
	for _, v := range body {
		v.write(&buf)
	}
	binary.Write(w, binary.BigEndian, uint16(buf.Len()))
	buf.WriteTo(w)
}

func ReadMsg(r io.Reader) (err error, c Cmd, msg interface{}) {
	var size uint16
	err = binary.Read(r, binary.BigEndian, &size)
	if err != nil {
		return
	}

	c = Cmd(readByte(r))
	size--

	typ := msgTypes[c]

	if typ == nil {
		err = errors.New(fmt.Sprintf("Cannot deserialize %s", c))
		return
	}

	v := reflect.New(typ).Elem()
	for f := 0; f < typ.NumField(); f++ {
		i := v.Field(f).Addr().Interface()
		w := i.(readable)
		size -= w.read(r)

		if size < 0 {
			err = errors.New(fmt.Sprintf("The %s message was too short to pOpulate all fields.", c))
			return
		}
	}

	if size > 0 {
		err = errors.New(fmt.Sprintf("The %s message was too long to pOpulate all fields.", c))
		return
	}

	msg = v.Interface()

	//fmt.Printf("<- %s %+v\n", c, msg)

	return
}

func WriteLogin(w io.Writer, nick Text, pass Text) {
	writeMessage(w, LOGIN, nick, pass)
}

func WriteLogout(w io.Writer) {
	writeMessage(w, LOGOUT)
}

func WriteStartGame(
	w io.Writer,
	numSeats Big,
	required Big,
	l LimitType,
	low Amount,
	high Amount) {

	writeMessage(
		w,
		START_GAME,
		Text("Test Table"),
		GT_TEXAS_HOLDEM,
		l,
		low,
		high,
		numSeats,     // seat count
		required,     // required players
		Big(14000),   // start delay
		Big(3600000), // player timeout
		Small(0),
	)
}

func WriteFold(w io.Writer, gid Id) {
	writeMessage(w, FOLD, gid)
}

func WriteSeatQuery(w io.Writer, gid Id) {
	writeMessage(w, SEAT_QUERY, gid)
}

func WriteGameQuery(
	w io.Writer,
	l LimitType,
	OpSeats Op,
	numSeats Small,
	OpJoined Op,
	joined Small,
	OpWaiting Op,
	waiting Small) {

	writeMessage(
		w,
		GAME_QUERY,
		GT_TEXAS_HOLDEM,
		l,
		OpSeats,
		numSeats,
		OpJoined,
		joined,
		OpWaiting,
		waiting)

}

func WriteJoin(w io.Writer, gid Id, seat Small, amt Amount) {
	writeMessage(w, JOIN, gid, seat, amt)
}

func WriteWatch(w io.Writer, gid Id) {
	writeMessage(w, WATCH, gid)
}

func WriteUnwatch(w io.Writer, gid Id) {
	writeMessage(w, UNWATCH, gid)
}

func WriteLeave(w io.Writer, gid Id) {
	writeMessage(w, LEAVE, gid)
}

func WriteRaise(w io.Writer, gid Id, raiseAmount Amount) {
	writeMessage(w, RAISE, gid, raiseAmount)
}

func WriteBalanceQuery(w io.Writer) {
	writeMessage(w, BALANCE_QUERY)
}

func WriteSitOut(w io.Writer, gid Id) {
	writeMessage(w, SIT_OUT, gid)
}

func WriteComeBack(w io.Writer, gid Id) {
	writeMessage(w, COME_BACK, gid)
}

func WriteChat(w io.Writer, msg Text) {
	writeMessage(w, CHAT, msg)
}

func WritePlayerQuery(w io.Writer, pid Id) {
	writeMessage(w, PLAYER_QUERY, pid)
}

type MsgGood struct {
	Cmd   Cmd
	Extra Big
}

type MsgBad struct {
	Cmd   Cmd
	Error Small
}

type MsgYouAre struct {
	Pid Id
}

type MsgYourGame struct {
	Gid Id
}

type MsgSeatInfo struct {
	Gid     Id
	SeatNum Small
	State   PlayerState
	Pid     Id
	InPlay  Amount
}

type MsgGameInfo struct {
	Gid       Id
	TableName Text
	GameType  GameType
	LimitType LimitType
	Low       Amount
	High      Amount
	NumSeats  Big
	Required  Big
	Joined    Big
	Waiting   Big
}

type MsgNotifyJoin struct {
	Gid    Id
	Pid    Id
	Seat   Small
	Amount Amount
}

type MsgNotifyCancelGame struct {
	Gid Id
}

type MsgNotifyLeave struct {
	Gid Id
	Pid Id
}

type MsgNotifyStartGame struct {
	Gid Id
}

type MsgNotifyButton struct {
	Gid    Id
	Button Small
}

type MsgNotifySb struct {
	Gid Id
	Sb  Small
}

type MsgBetReq struct {
	Gid        Id
	CallAmount Amount
	RaiseMin   Amount
	RaiseMax   Amount
}

type MsgBalance struct {
	Balance Amount
	InPlay  Amount
}

type MsgNotifyRaise struct {
	Gid         Id
	Pid         Id
	RaiseAmount Amount
	CallAmount  Amount
}

type MsgNotifyBb struct {
	Gid Id
	Bb  Small
}

type MsgNotifyDraw struct {
	Gid  Id
	Pid  Id
	Seq  Seq
	Suit Suit
}

type MsgNotifyShared struct {
	Gid  Id
	Seq  Seq
	Suit Suit
}

type MsgNotifyHand struct {
	Gid   Id
	Pid   Id
	Rank  Small
	Face1 Small
	Face2 Small
}

type MsgNotifyEndGame struct {
	Gid Id
}

type MsgNotifyChat struct {
	Gid Id
	Pid Id
	Msg Text
}

type MsgGameStage struct {
	Gid   Id
	Stage GameStage
}

type MsgShowCards struct {
	Gid   Id
	Pid   Id
	Cards Cards
}

type MsgNotifyWin struct {
	Gid    Id
	Pid    Id
	Amount Amount
}

type MsgPlayerInfo struct {
	Pid         Id
	TotalInPlay Amount
	Nick        Text
	Location    Text
}

func (c GameStage) String() string {
	switch c {
	case GS_BLINDS:
		return "Blinds"
	case GS_FLOP:
		return "Flop"
	case GS_PREFLOP:
		return "Pre-flop"
	case GS_RIVER:
		return "River"
	case GS_SHOWDOWN:
		return "Showdown"
	case GS_TURN:
		return "Turn"
	case GS_DELAYED_START:
		return "Delayed start"
	}
	return "Unknown game stage"
}

func (c Cmd) String() string {
	switch c {
	case GOOD:
		return "GOOD"
	case BAD:
		return "BAD"
	case LOGIN:
		return "LOGIN"
	case START_GAME:
		return "START_GAME"
	case YOU_ARE:
		return "YOU_ARE"
	case YOUR_GAME:
		return "YOUR_GAME"
	case SEAT_INFO:
		return "SEAT_INFO"
	case GAME_INFO:
		return "GAME_INFO"
	case NOTIFY_JOIN:
		return "NOTIFY_JOIN"
	case NOTIFY_CANCEL_GAME:
		return "NOTIFY_CANCEL_GAME"
	case LOGOUT:
		return "LOGOUT"
	case GAME_QUERY:
		return "GAME_QUERY"
	case UNWATCH:
		return "UNWATCH"
	case LEAVE:
		return "LEAVE"
	case NOTIFY_LEAVE:
		return "NOTIFY_LEAVE"
	case WATCH:
		return "WATCH"
	case JOIN:
		return "JOIN"
	case NOTIFY_START_GAME:
		return "NOTIFY_START_GAME"
	case NOTIFY_BUTTON:
		return "NOTIFY_BUTTON"
	case NOTIFY_SB:
		return "NOTIFY_SB"
	case BET_REQ:
		return "BET_REQ"
	case RAISE:
		return "RAISE"
	case SEAT_QUERY:
		return "SEAT_QUERY"
	case BALANCE_QUERY:
		return "BALANCE_QUERY"
	case BALANCE:
		return "BALANCE"
	case FOLD:
		return "FOLD"
	case NOTIFY_RAISE:
		return "NOTIFY_RAISE"
	case NOTIFY_BB:
		return "NOTIFY_BB"
	case NOTIFY_DRAW:
		return "NOTIFY_DRAW"
	case NOTIFY_SHARED:
		return "NOTIFY_SHARED"
	case NOTIFY_HAND:
		return "NOTIFY_HAND"
	case NOTIFY_END_GAME:
		return "NOTIFY_END_GAME"
	case SIT_OUT:
		return "SIT_OUT"
	case COME_BACK:
		return "COME_BACK"
	case CHAT:
		return "CHAT"
	case NOTIFY_CHAT:
		return "NOTIFY_CHAT"
	case GAME_STAGE:
		return "GAME_STAGE"
	case SHOW_CARDS:
		return "SHOW_CARDS"
	case NOTIFY_WIN:
		return "NOTIFY_WIN"
	case PLAYER_QUERY:
		return "PLAYER_QUERY"
	case PLAYER_INFO:
		return "PLAYER_INFO"
	}
	return fmt.Sprintf("Unknown Command (%d)", uint8(c))
}

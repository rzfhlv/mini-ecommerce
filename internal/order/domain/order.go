package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusPaid      OrderStatus = "PAID"
	OrderStatusCancelled OrderStatus = "CANCELLED"
)

var (
	ErrOrderEmpty              = errors.New("order: cannot create order with no items")
	ErrInvalidStatusChange     = errors.New("order: invalid status transition")
	ErrDiscountExceedsSubtotal = errors.New("discount: discount amount cannot exceed subtotal")
)

type OrderItem struct {
	id, productID uuid.UUID
	productName   string
	price         Money
	quantity      int
}

func (i OrderItem) Subtotal() Money      { return i.price.Multiply(i.quantity) }
func (i OrderItem) ID() uuid.UUID        { return i.id }
func (i OrderItem) ProductID() uuid.UUID { return i.productID }
func (i OrderItem) ProductName() string  { return i.productName }
func (i OrderItem) Price() Money         { return i.price }
func (i OrderItem) Quantity() int        { return i.quantity }

func ReconstructOrderItem(id, productID uuid.UUID, productName string, price Money, quantity int) OrderItem {
	return OrderItem{id: id, productID: productID, productName: productName, price: price, quantity: quantity}
}

type Order struct {
	id, userID       uuid.UUID
	items            []OrderItem
	subtotal         Money  // total item SEBELUM diskon/fee -- tidak berubah setelah order dibuat
	discountAmount   Money
	discountLabel    string
	adminFeeAmount   Money  // BARU -- biaya admin, MENAMBAH total (kebalikan dari discount)
	adminFeeLabel    string
	totalAmount      Money  // subtotal - discountAmount + adminFeeAmount -- lihat recomputeTotal()
	status           OrderStatus
	paymentReference string
	createdAt        time.Time
	paidAt           *time.Time

	domainEvents []DomainEvent
}

func NewOrder(userID uuid.UUID, items []OrderItem, currency Currency) (*Order, error) {
	if len(items) == 0 {
		return nil, ErrOrderEmpty
	}
	subtotal := ZeroMoney(currency)
	for _, item := range items {
		var err error
		subtotal, err = subtotal.Add(item.Subtotal())
		if err != nil {
			return nil, err
		}
	}
	order := &Order{
		id: uuid.New(), userID: userID, items: items,
		subtotal: subtotal, discountAmount: ZeroMoney(currency), adminFeeAmount: ZeroMoney(currency),
		totalAmount: subtotal, status: OrderStatusPending, createdAt: time.Now(),
	}
	order.raise(OrderPlacedEvent{OrderID: order.id, UserID: order.userID, Total: order.totalAmount, OccuredAt: order.createdAt})
	return order, nil
}

func ReconstructOrder(id, userID uuid.UUID, items []OrderItem, subtotal, discountAmount Money, discountLabel string, adminFeeAmount Money, adminFeeLabel string, totalAmount Money, status OrderStatus, paymentReference string, createdAt time.Time, paidAt *time.Time) *Order {
	return &Order{
		id: id, userID: userID, items: items, subtotal: subtotal,
		discountAmount: discountAmount, discountLabel: discountLabel,
		adminFeeAmount: adminFeeAmount, adminFeeLabel: adminFeeLabel,
		totalAmount: totalAmount, status: status, paymentReference: paymentReference,
		createdAt: createdAt, paidAt: paidAt,
	}
}

// recomputeTotal -- SATU-SATUNYA tempat rumus total dihitung.
// Dipanggil oleh ApplyDiscount DAN ApplyAdminFee supaya urutan
// pemanggilan (diskon dulu vs fee dulu) tidak pernah menghasilkan
// hasil yang beda -- keduanya sama-sama menghitung ULANG dari
// subtotal, bukan menambah/mengurangi dari totalAmount sebelumnya.
func (o *Order) recomputeTotal() {
	afterDiscount := o.subtotal.Amount() - o.discountAmount.Amount()
	total := afterDiscount + o.adminFeeAmount.Amount()
	o.totalAmount, _ = NewMoney(total, o.subtotal.Currency())
}

func (o *Order) ApplyDiscount(policy DiscountPolicy) error {
	if o.status != OrderStatusPending {
		return ErrInvalidStatusChange
	}
	discount, err := policy.Apply(o.subtotal)
	if err != nil {
		return err
	}
	if discount.Amount() > o.subtotal.Amount() {
		return ErrDiscountExceedsSubtotal
	}
	o.discountAmount = discount
	o.discountLabel = policy.Describe()
	o.recomputeTotal()
	if discount.Amount() > 0 {
		o.raise(OrderDiscountAppliedEvent{OrderID: o.id, DiscountAmount: discount, DiscountLabel: policy.Describe()})
	}
	return nil
}

// ApplyAdminFee -- pola SAMA PERSIS ApplyDiscount, tapi menambah
// nilai ke total, bukan mengurangi.
func (o *Order) ApplyAdminFee(policy AdminFeePolicy) error {
	if o.status != OrderStatusPending {
		return ErrInvalidStatusChange
	}
	fee, err := policy.Apply(o.subtotal)
	if err != nil {
		return err
	}
	o.adminFeeAmount = fee
	o.adminFeeLabel = policy.Describe()
	o.recomputeTotal()
	if fee.Amount() > 0 {
		o.raise(OrderAdminFeeAppliedEvent{OrderID: o.id, FeeAmount: fee, FeeLabel: policy.Describe()})
	}
	return nil
}

func (o *Order) MarkAsPaid() error {
	if o.status != OrderStatusPending {
		return ErrInvalidStatusChange
	}
	now := time.Now()
	o.status = OrderStatusPaid
	o.paidAt = &now
	o.raise(OrderPaidEvent{OrderID: o.id, OccuredAt: now})
	return nil
}

func (o *Order) Cancel() error {
	if o.status != OrderStatusPending {
		return ErrInvalidStatusChange
	}
	o.status = OrderStatusCancelled
	return nil
}

func (o *Order) SetPaymentReference(ref string) error {
	if o.status != OrderStatusPending {
		return ErrInvalidStatusChange
	}
	o.paymentReference = ref
	return nil
}

func (o *Order) ID() uuid.UUID            { return o.id }
func (o *Order) UserID() uuid.UUID        { return o.userID }
func (o *Order) Status() OrderStatus      { return o.status }
func (o *Order) TotalAmount() Money       { return o.totalAmount }
func (o *Order) Items() []OrderItem       { return o.items }
func (o *Order) Subtotal() Money          { return o.subtotal }
func (o *Order) DiscountAmount() Money    { return o.discountAmount }
func (o *Order) DiscountLabel() string    { return o.discountLabel }
func (o *Order) AdminFeeAmount() Money    { return o.adminFeeAmount }
func (o *Order) AdminFeeLabel() string    { return o.adminFeeLabel }
func (o *Order) PaymentReference() string { return o.paymentReference }
func (o *Order) CreatedAt() time.Time     { return o.createdAt }
func (o *Order) PaidAt() *time.Time       { return o.paidAt }

func (o *Order) raise(e DomainEvent)      { o.domainEvents = append(o.domainEvents, e) }
func (o *Order) PullEvents() []DomainEvent { e := o.domainEvents; o.domainEvents = nil; return e }

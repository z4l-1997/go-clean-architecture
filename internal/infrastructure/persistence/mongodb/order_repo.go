// Package mongodb chứa các MongoDB repository implementations
package mongodb

import (
	"context"
	"errors"
	"time"

	"restaurant_project/internal/domain/entity"
	"restaurant_project/internal/domain/repository"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// orderItemDocument là struct mapping cho OrderItem trong MongoDB
type orderItemDocument struct {
	MonAnID   string `bson:"mon_an_id"`
	TenMon    string `bson:"ten_mon"`
	SoLuong   int    `bson:"so_luong"`
	DonGia    int64  `bson:"don_gia"`
	GhiChu    string `bson:"ghi_chu,omitempty"`
	ThanhTien int64  `bson:"thanh_tien"`
}

// orderDocument là struct mapping với MongoDB document
type orderDocument struct {
	ID                string              `bson:"_id"`
	KhachHangID       string              `bson:"khach_hang_id,omitempty"`
	NhanVienID        string              `bson:"nhan_vien_id,omitempty"`
	DauBepID          string              `bson:"dau_bep_id,omitempty"`
	SoBan             int                 `bson:"so_ban,omitempty"`
	LoaiOrder         string              `bson:"loai_order"`
	TrangThai         string              `bson:"trang_thai"`
	Items             []orderItemDocument `bson:"items"`
	TongTien          int64               `bson:"tong_tien"`
	GiamGia           int64               `bson:"giam_gia"`
	TienThanhToan     int64               `bson:"tien_thanh_toan"`
	GhiChu            string              `bson:"ghi_chu,omitempty"`
	DiaChiGiao        string              `bson:"dia_chi_giao,omitempty"`
	ThoiGianDat       time.Time           `bson:"thoi_gian_dat"`
	ThoiGianCapNhat   time.Time           `bson:"thoi_gian_cap_nhat"`
	ThoiGianHoanThanh *time.Time          `bson:"thoi_gian_hoan_thanh,omitempty"`
}

// toEntity chuyển từ document sang entity
func (d *orderDocument) toEntity() *entity.Order {
	items := make([]entity.OrderItem, len(d.Items))
	for i, item := range d.Items {
		items[i] = entity.OrderItem{
			MonAnID:   item.MonAnID,
			TenMon:    item.TenMon,
			SoLuong:   item.SoLuong,
			DonGia:    item.DonGia,
			GhiChu:    item.GhiChu,
			ThanhTien: item.ThanhTien,
		}
	}

	return &entity.Order{
		ID:                d.ID,
		KhachHangID:       d.KhachHangID,
		NhanVienID:        d.NhanVienID,
		DauBepID:          d.DauBepID,
		SoBan:             d.SoBan,
		LoaiOrder:         entity.LoaiOrder(d.LoaiOrder),
		TrangThai:         entity.TrangThaiOrder(d.TrangThai),
		Items:             items,
		TongTien:          d.TongTien,
		GiamGia:           d.GiamGia,
		TienThanhToan:     d.TienThanhToan,
		GhiChu:            d.GhiChu,
		DiaChiGiao:        d.DiaChiGiao,
		ThoiGianDat:       d.ThoiGianDat,
		ThoiGianCapNhat:   d.ThoiGianCapNhat,
		ThoiGianHoanThanh: d.ThoiGianHoanThanh,
	}
}

// toOrderDocument chuyển từ entity sang document
func toOrderDocument(o *entity.Order) *orderDocument {
	items := make([]orderItemDocument, len(o.Items))
	for i, item := range o.Items {
		items[i] = orderItemDocument{
			MonAnID:   item.MonAnID,
			TenMon:    item.TenMon,
			SoLuong:   item.SoLuong,
			DonGia:    item.DonGia,
			GhiChu:    item.GhiChu,
			ThanhTien: item.ThanhTien,
		}
	}

	return &orderDocument{
		ID:                o.ID,
		KhachHangID:       o.KhachHangID,
		NhanVienID:        o.NhanVienID,
		DauBepID:          o.DauBepID,
		SoBan:             o.SoBan,
		LoaiOrder:         string(o.LoaiOrder),
		TrangThai:         string(o.TrangThai),
		Items:             items,
		TongTien:          o.TongTien,
		GiamGia:           o.GiamGia,
		TienThanhToan:     o.TienThanhToan,
		GhiChu:            o.GhiChu,
		DiaChiGiao:        o.DiaChiGiao,
		ThoiGianDat:       o.ThoiGianDat,
		ThoiGianCapNhat:   o.ThoiGianCapNhat,
		ThoiGianHoanThanh: o.ThoiGianHoanThanh,
	}
}

// OrderMongoRepo là implementation của IOrderRepository sử dụng MongoDB
type OrderMongoRepo struct {
	collection *mongo.Collection
}

// NewOrderMongoRepo tạo mới OrderMongoRepo
func NewOrderMongoRepo(db *mongo.Database) *OrderMongoRepo {
	return &OrderMongoRepo{
		collection: db.Collection("orders"),
	}
}

// Verify interface implementation at compile time
var _ repository.IOrderRepository = (*OrderMongoRepo)(nil)

// FindByID tìm order theo ID
func (r *OrderMongoRepo) FindByID(ctx context.Context, id string) (*entity.Order, error) {
	var doc orderDocument
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return doc.toEntity(), nil
}

// FindAll lấy tất cả orders
func (r *OrderMongoRepo) FindAll(ctx context.Context) ([]*entity.Order, error) {
	opts := options.Find().SetSort(bson.M{"thoi_gian_dat": -1})
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []*entity.Order
	for cursor.Next(ctx) {
		var doc orderDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		list = append(list, doc.toEntity())
	}

	return list, cursor.Err()
}

// FindByKhachHangID lấy orders của một khách hàng
func (r *OrderMongoRepo) FindByKhachHangID(ctx context.Context, khachHangID string) ([]*entity.Order, error) {
	opts := options.Find().SetSort(bson.M{"thoi_gian_dat": -1})
	cursor, err := r.collection.Find(ctx, bson.M{"khach_hang_id": khachHangID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []*entity.Order
	for cursor.Next(ctx) {
		var doc orderDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		list = append(list, doc.toEntity())
	}

	return list, cursor.Err()
}

// FindByTrangThai lấy orders theo trạng thái
func (r *OrderMongoRepo) FindByTrangThai(ctx context.Context, trangThai entity.TrangThaiOrder) ([]*entity.Order, error) {
	opts := options.Find().SetSort(bson.M{"thoi_gian_dat": -1})
	cursor, err := r.collection.Find(ctx, bson.M{"trang_thai": string(trangThai)}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []*entity.Order
	for cursor.Next(ctx) {
		var doc orderDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		list = append(list, doc.toEntity())
	}

	return list, cursor.Err()
}

// FindByDauBepID lấy orders được gán cho một đầu bếp
func (r *OrderMongoRepo) FindByDauBepID(ctx context.Context, dauBepID string) ([]*entity.Order, error) {
	opts := options.Find().SetSort(bson.M{"thoi_gian_dat": -1})
	cursor, err := r.collection.Find(ctx, bson.M{"dau_bep_id": dauBepID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []*entity.Order
	for cursor.Next(ctx) {
		var doc orderDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		list = append(list, doc.toEntity())
	}

	return list, cursor.Err()
}

// FindByThoiGian lấy orders trong khoảng thời gian
func (r *OrderMongoRepo) FindByThoiGian(ctx context.Context, from, to time.Time) ([]*entity.Order, error) {
	filter := bson.M{
		"thoi_gian_dat": bson.M{
			"$gte": from,
			"$lte": to,
		},
	}
	opts := options.Find().SetSort(bson.M{"thoi_gian_dat": -1})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []*entity.Order
	for cursor.Next(ctx) {
		var doc orderDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		list = append(list, doc.toEntity())
	}

	return list, cursor.Err()
}

// FindPending lấy các orders đang chờ xử lý
func (r *OrderMongoRepo) FindPending(ctx context.Context) ([]*entity.Order, error) {
	filter := bson.M{
		"trang_thai": bson.M{
			"$in": []string{
				string(entity.OrderMoi),
				string(entity.OrderDaXacNhan),
				string(entity.OrderDangNau),
			},
		},
	}
	opts := options.Find().SetSort(bson.M{"thoi_gian_dat": 1}) // FIFO - đơn cũ trước
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []*entity.Order
	for cursor.Next(ctx) {
		var doc orderDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		list = append(list, doc.toEntity())
	}

	return list, cursor.Err()
}

// Save lưu order mới hoặc cập nhật
func (r *OrderMongoRepo) Save(ctx context.Context, order *entity.Order) error {
	doc := toOrderDocument(order)

	opts := options.Replace().SetUpsert(true)
	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": order.ID}, doc, opts)

	return err
}

// Delete xóa order theo ID
func (r *OrderMongoRepo) Delete(ctx context.Context, id string) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("không tìm thấy order để xóa")
	}

	return nil
}

// UpdateTrangThai cập nhật trạng thái order
func (r *OrderMongoRepo) UpdateTrangThai(ctx context.Context, id string, trangThai entity.TrangThaiOrder) error {
	update := bson.M{
		"$set": bson.M{
			"trang_thai":       string(trangThai),
			"thoi_gian_cap_nhat": time.Now(),
		},
	}

	// Nếu hoàn thành, cập nhật thời gian hoàn thành
	if trangThai == entity.OrderHoanThanh {
		now := time.Now()
		update["$set"].(bson.M)["thoi_gian_hoan_thanh"] = now
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("không tìm thấy order để cập nhật")
	}

	return nil
}

// Count đếm tổng số orders
func (r *OrderMongoRepo) Count(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{})
}

// CountByTrangThai đếm orders theo trạng thái
func (r *OrderMongoRepo) CountByTrangThai(ctx context.Context, trangThai entity.TrangThaiOrder) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{"trang_thai": string(trangThai)})
}

// TinhDoanhThu tính doanh thu trong khoảng thời gian
func (r *OrderMongoRepo) TinhDoanhThu(ctx context.Context, from, to time.Time) (int64, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"trang_thai": string(entity.OrderHoanThanh),
				"thoi_gian_hoan_thanh": bson.M{
					"$gte": from,
					"$lte": to,
				},
			},
		},
		{
			"$group": bson.M{
				"_id":        nil,
				"tong_doanh_thu": bson.M{"$sum": "$tien_thanh_toan"},
			},
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var result struct {
		TongDoanhThu int64 `bson:"tong_doanh_thu"`
	}

	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return 0, err
		}
	}

	return result.TongDoanhThu, nil
}

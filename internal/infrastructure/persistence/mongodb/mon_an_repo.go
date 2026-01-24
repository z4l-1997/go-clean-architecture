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

// monAnDocument là struct mapping với MongoDB document
type monAnDocument struct {
	ID          string    `bson:"_id"`
	Ten         string    `bson:"ten"`
	Gia         int64     `bson:"gia"`
	MoTa        string    `bson:"mo_ta"`
	ConHang     bool      `bson:"con_hang"`
	GiamGia     int       `bson:"giam_gia"`
	NgayTao     time.Time `bson:"ngay_tao"`
	NgayCapNhat time.Time `bson:"ngay_cap_nhat"`
}

// toEntity chuyển từ document sang entity
func (d *monAnDocument) toEntity() *entity.MonAn {
	return &entity.MonAn{
		ID:          d.ID,
		Ten:         d.Ten,
		Gia:         d.Gia,
		MoTa:        d.MoTa,
		ConHang:     d.ConHang,
		GiamGia:     d.GiamGia,
		NgayTao:     d.NgayTao,
		NgayCapNhat: d.NgayCapNhat,
	}
}

// toDocument chuyển từ entity sang document
func toMonAnDocument(m *entity.MonAn) *monAnDocument {
	return &monAnDocument{
		ID:          m.ID,
		Ten:         m.Ten,
		Gia:         m.Gia,
		MoTa:        m.MoTa,
		ConHang:     m.ConHang,
		GiamGia:     m.GiamGia,
		NgayTao:     m.NgayTao,
		NgayCapNhat: m.NgayCapNhat,
	}
}

// MonAnMongoRepo là implementation của IMonAnRepository sử dụng MongoDB
type MonAnMongoRepo struct {
	collection *mongo.Collection
}

// NewMonAnMongoRepo tạo mới MonAnMongoRepo
func NewMonAnMongoRepo(db *mongo.Database) *MonAnMongoRepo {
	return &MonAnMongoRepo{
		collection: db.Collection("mon_an"),
	}
}

// Verify interface implementation at compile time
var _ repository.IMonAnRepository = (*MonAnMongoRepo)(nil)

// FindByID tìm món ăn theo ID
func (r *MonAnMongoRepo) FindByID(ctx context.Context, id string) (*entity.MonAn, error) {
	var doc monAnDocument
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return doc.toEntity(), nil
}

// FindAll lấy tất cả món ăn
func (r *MonAnMongoRepo) FindAll(ctx context.Context) ([]*entity.MonAn, error) {
	opts := options.Find().SetSort(bson.M{"ngay_tao": -1})
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []*entity.MonAn
	for cursor.Next(ctx) {
		var doc monAnDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		list = append(list, doc.toEntity())
	}

	return list, cursor.Err()
}

// FindByConHang lấy các món theo trạng thái còn hàng
func (r *MonAnMongoRepo) FindByConHang(ctx context.Context, conHang bool) ([]*entity.MonAn, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"con_hang": conHang})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []*entity.MonAn
	for cursor.Next(ctx) {
		var doc monAnDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		list = append(list, doc.toEntity())
	}

	return list, cursor.Err()
}

// Save lưu món ăn mới hoặc cập nhật
func (r *MonAnMongoRepo) Save(ctx context.Context, mon *entity.MonAn) error {
	doc := toMonAnDocument(mon)

	opts := options.Replace().SetUpsert(true)
	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": mon.ID}, doc, opts)

	return err
}

// Delete xóa món ăn theo ID
func (r *MonAnMongoRepo) Delete(ctx context.Context, id string) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("không tìm thấy món ăn để xóa")
	}

	return nil
}

// Count đếm tổng số món ăn
func (r *MonAnMongoRepo) Count(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{})
}

// Package persistence chứa các implementation cho Repository
// Đây là "kho lưu trữ" - nơi thực sự lưu/lấy dữ liệu
package memory

import (
	"context"
	"errors"
	"sync"

	"restaurant_project/internal/domain/entity"
	"restaurant_project/internal/domain/repository"
)

// MonAnMemoryRepo là implementation của IMonAnRepository sử dụng memory
//
// TẠI SAO DÙNG MEMORY TRƯỚC?
// 1. Đơn giản, dễ hiểu - phù hợp cho người mới học
// 2. Không cần setup database
// 3. Sau này có thể dễ dàng thay bằng MySQL, MongoDB mà KHÔNG cần sửa code khác
//
// CLEAN ARCHITECTURE LỢI ÍCH:
// - UseCase không biết dữ liệu lưu ở đâu
// - Chỉ cần implement interface IMonAnRepository
// - Có thể swap implementation bất cứ lúc nào
type MonAnMemoryRepo struct {
	data  map[string]*entity.MonAn // Map lưu trữ: key = ID, value = MonAn
	mutex sync.RWMutex             // Mutex để thread-safe
}

// NewMonAnMemoryRepo tạo mới MonAnMemoryRepo
func NewMonAnMemoryRepo() *MonAnMemoryRepo {
	return &MonAnMemoryRepo{
		data: make(map[string]*entity.MonAn),
	}
}

// Verify interface implementation at compile time
// Đây là Go idiom để đảm bảo struct implement đúng interface
var _ repository.IMonAnRepository = (*MonAnMemoryRepo)(nil)

// FindByID tìm món ăn theo ID
func (r *MonAnMemoryRepo) FindByID(ctx context.Context, id string) (*entity.MonAn, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	mon, exists := r.data[id]
	if !exists {
		return nil, nil // Không tìm thấy, trả về nil (không phải error)
	}

	// Trả về copy để tránh modification từ bên ngoài
	return r.copyMonAn(mon), nil
}

// FindAll lấy tất cả món ăn
func (r *MonAnMemoryRepo) FindAll(ctx context.Context) ([]*entity.MonAn, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	result := make([]*entity.MonAn, 0, len(r.data))
	for _, mon := range r.data {
		result = append(result, r.copyMonAn(mon))
	}

	return result, nil
}

// FindByConHang lấy các món theo trạng thái còn hàng
func (r *MonAnMemoryRepo) FindByConHang(ctx context.Context, conHang bool) ([]*entity.MonAn, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	result := make([]*entity.MonAn, 0)
	for _, mon := range r.data {
		if mon.ConHang == conHang {
			result = append(result, r.copyMonAn(mon))
		}
	}

	return result, nil
}

// Save lưu món ăn (insert hoặc update)
func (r *MonAnMemoryRepo) Save(ctx context.Context, mon *entity.MonAn) error {
	if mon == nil {
		return errors.New("món ăn không được nil")
	}

	if mon.ID == "" {
		return errors.New("ID món ăn không được để trống")
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Lưu copy để tránh modification từ bên ngoài
	r.data[mon.ID] = r.copyMonAn(mon)

	return nil
}

// Delete xóa món ăn theo ID
func (r *MonAnMemoryRepo) Delete(ctx context.Context, id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.data[id]; !exists {
		return errors.New("không tìm thấy món ăn để xóa")
	}

	delete(r.data, id)
	return nil
}

// Count đếm tổng số món ăn
func (r *MonAnMemoryRepo) Count(ctx context.Context) (int64, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return int64(len(r.data)), nil
}

// ============================================
// HELPER FUNCTIONS
// ============================================

// copyMonAn tạo bản copy của MonAn
// Điều này quan trọng để tránh "shared state" bugs
// VD: Nếu trả về pointer trực tiếp, caller có thể modify data trong repo
func (r *MonAnMemoryRepo) copyMonAn(mon *entity.MonAn) *entity.MonAn {
	return &entity.MonAn{
		ID:          mon.ID,
		Ten:         mon.Ten,
		Gia:         mon.Gia,
		MoTa:        mon.MoTa,
		ConHang:     mon.ConHang,
		GiamGia:     mon.GiamGia,
		NgayTao:     mon.NgayTao,
		NgayCapNhat: mon.NgayCapNhat,
	}
}

// ============================================
// SEED DATA (Optional - để test)
// ============================================

// SeedSampleData thêm dữ liệu mẫu để test
func (r *MonAnMemoryRepo) SeedSampleData() {
	sampleData := []*entity.MonAn{
		{
			ID:      "1",
			Ten:     "Phở tái",
			Gia:     50000,
			MoTa:    "Phở bò tái thơm ngon",
			ConHang: true,
			GiamGia: 0,
		},
		{
			ID:      "2",
			Ten:     "Bún bò Huế",
			Gia:     55000,
			MoTa:    "Bún bò cay nồng đặc trưng Huế",
			ConHang: true,
			GiamGia: 10,
		},
		{
			ID:      "3",
			Ten:     "Cơm tấm sườn",
			Gia:     45000,
			MoTa:    "Cơm tấm với sườn nướng",
			ConHang: false,
			GiamGia: 0,
		},
	}

	for _, mon := range sampleData {
		r.data[mon.ID] = mon
	}
}

package models

import (
	"errors"
	"html"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

type Favorite struct {
	ID        uint64    `gorm:"primary_key;auto_increment" json:"id"`
	Title     string    `gorm:"size:255;not null;unique" json:"title"`
	User      User      `json:"user"`
	UserID    uint32    `gorm:"not null" json:"user_id"`
	Watched   bool      `gorm:"type:bool;default:false"`
	WantWatch bool      `gorm:"type:bool;default:false"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (f *Favorite) Prepare() {
	f.ID = 0
	f.Title = html.EscapeString(strings.TrimSpace(f.Title))
	f.User = User{}
	f.Watched = false
	f.WantWatch = false
	f.CreatedAt = time.Now()
	f.UpdatedAt = time.Now()
}

func (f *Favorite) Validate() error {

	if f.Title == "" {
		return errors.New("required title")
	}
	if f.UserID < 1 {
		return errors.New("required user")
	}
	return nil
}

func (f *Favorite) SaveFavorite(db *gorm.DB) (*Favorite, error) {
	var err error
	err = db.Debug().Model(&Favorite{}).Create(&f).Error
	if err != nil {
		return &Favorite{}, err
	}
	if f.ID != 0 {
		err = db.Debug().Model(&User{}).Where("id = ?", f.UserID).Take(&f.User).Error
		if err != nil {
			return &Favorite{}, err
		}
	}
	return f, nil
}

func (f *Favorite) FindAllFavorites(db *gorm.DB) (*[]Favorite, error) {
	var err error
	favorites := []Favorite{}
	err = db.Debug().Model(&Favorite{}).Limit(100).Find(&favorites).Error
	if err != nil {
		return &[]Favorite{}, err
	}
	if len(favorites) > 0 {
		for i, _ := range favorites {
			err := db.Debug().Model(&User{}).Where("id = ?", favorites[i].UserID).Take(&favorites[i].User).Error
			if err != nil {
				return &[]Favorite{}, err
			}
		}
	}
	return &favorites, nil
}

func (f *Favorite) FindFavoriteByID(db *gorm.DB, fid uint64) (*Favorite, error) {
	var err error
	err = db.Debug().Model(&Favorite{}).Where("id = ?", fid).Take(&f).Error
	if err != nil {
		return &Favorite{}, err
	}
	if f.ID != 0 {
		err = db.Debug().Model(&User{}).Where("id = ?", f.UserID).Take(&f.User).Error
		if err != nil {
			return &Favorite{}, err
		}
	}
	return f, nil
}

func (f *Favorite) UpdateAFavorite(db *gorm.DB) (*Favorite, error) {

	var err error

	err = db.Debug().Model(&Favorite{}).Where("id = ?", f.ID).Updates(Favorite{Title: f.Title, UpdatedAt: time.Now()}).Error
	if err != nil {
		return &Favorite{}, err
	}
	if f.ID != 0 {
		err = db.Debug().Model(&User{}).Where("id = ?", f.UserID).Take(&f.User).Error
		if err != nil {
			return &Favorite{}, err
		}
	}
	return f, nil
}

func (f *Favorite) DeleteAFavorite(db *gorm.DB, fid uint64, uid uint32) (int64, error) {

	db = db.Debug().Model(&Favorite{}).Where("id = ? and user_id = ?", fid, uid).Take(&Favorite{}).Delete(&Favorite{})

	if db.Error != nil {
		if gorm.IsRecordNotFoundError(db.Error) {
			return 0, errors.New("Favorite not found")
		}
		return 0, db.Error
	}
	return db.RowsAffected, nil
}

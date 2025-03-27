package service

import (
	"errors"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
	"payment-service/dto/request"
	"payment-service/model"
	"payment-service/repository"
	"time"
)

type PaymentServiceInterface interface {
	ProcessPayment(request request.PaymentRequest) error
}

type PaymentService struct {
	db                    *gorm.DB
	accountRepository     repository.AccountRepositoryInterface
	transactionRepository repository.TransactionRepositoryInterface
	redisService          RedisServiceInterface
}

func NewPaymentService(
	db *gorm.DB,
	accountRepo repository.AccountRepositoryInterface,
	transactionRepo repository.TransactionRepositoryInterface,
	redisService RedisServiceInterface,
) *PaymentService {
	return &PaymentService{
		db:                    db,
		accountRepository:     accountRepo,
		transactionRepository: transactionRepo,
		redisService:          redisService,
	}
}

func (ps *PaymentService) ProcessPayment(req request.PaymentRequest) error {
	valid, err := ps.redisService.IdempotencyValidation(req.UUID)
	if err != nil {
		return err
	}
	if !valid {
		return errors.New("duplicate or invalid request")
	}

	tx := ps.getDbConnection()
	if err := ps.accountRepository.DeductAmount(tx, req.AccountID, req.Amount); err != nil {
		tx.Rollback()
		return err
	}

	newTxn := &model.Transaction{
		TransactionId: uuid.NewV4().String(),
		ProductId:     req.ProductId,
		Amount:        req.Amount,
		CreateDate:    time.Now().UTC(),
		RequestId:     req.RequestID,
	}
	if err := ps.transactionRepository.Insert(tx, newTxn); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

func (ps *PaymentService) getDbConnection() *gorm.DB {
	tx := ps.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	return tx
}

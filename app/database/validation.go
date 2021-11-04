package database

import (
	"database/sql"
	"fmt"
	scCsv "site-controller-data-update-to-mysql/app/csv"
	"site-controller-data-update-to-mysql/app/helper"
	"site-controller-data-update-to-mysql/app/models"
	"time"

	"context"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func checkInTime(reservation *scCsv.ReservationData) (time.Time, error) {
	if len(reservation.CheckInTime) > 0 {
		stayDateFrom, err := time.Parse("20060102 15:04", reservation.StayDateFrom+" "+reservation.CheckInTime)
		if err != nil {
			return time.Time{}, fmt.Errorf("failed to parse checkInTime: %v\n", err)
		}
		return stayDateFrom, nil
	} else {
		stayDateFrom, err := time.Parse("20060102", reservation.StayDateFrom)
		if err != nil {
			return time.Time{}, fmt.Errorf("failed to parse checkInTime: %v\n", err)
		}
		return stayDateFrom, nil
	}
}

func checkChild(reservation *scCsv.ReservationData) int8 {
	numChild := reservation.NumberOfGuestsChildA + reservation.NumberOfGuestsChildB + reservation.NumberOfGuestsChildC + reservation.NumberOfGuestsChildD
	if numChild > 0 {
		return 1 //有
	} else {
		return 0
	}
}

func checkPaymentMethod(paymentMethodName string, ctx context.Context, tx *sql.Tx) (int, error) {
	if paymentMethodName != "" {
		id, err := getPaymentMethod(paymentMethodName, ctx, tx)
		if err != nil {
			logging.Info(fmt.Sprintf("failed to get payment method error: %v", err), nil)
		}
		if id == 0 {
			id, err = insertPaymentMethod(paymentMethodName, ctx, tx)
			if err != nil {
				return 0, err
			}
			return id, nil
		}
		return id, nil
	}
	id, err := getPaymentMethod("指定なし", ctx, tx)
	if err != nil {
		logging.Info(fmt.Sprintf("failed to get payment method id correspond to '指定なし' error: %v", err), nil)
	}
	return id, nil
}

func getPaymentMethod(paymentMethodName string, ctx context.Context, tx *sql.Tx) (int, error) {
	record, err := models.PaymentMethodMasters(
		qm.Where(models.PaymentMethodMasterColumns.PaymentMethodName+" = ?", paymentMethodName),
	).One(ctx, tx)
	if err != nil {
		return 0, err
	}
	if record == nil {
		return 0, fmt.Errorf("%s not found", paymentMethodName)
	}
	return record.PaymentMethodID, nil
}

func insertPaymentMethod(paymentMethodName string, ctx context.Context, tx *sql.Tx) (int, error) {
	newPaymentMethodMaster := models.PaymentMethodMaster{
		//ReservationMethodID:   int,
		PaymentMethodName: null.StringFrom(paymentMethodName),
	}
	if err := newPaymentMethodMaster.Insert(ctx, tx, boil.Infer()); err != nil {
		return 0, err
	}
	return newPaymentMethodMaster.PaymentMethodID, nil
}

func checkReservationMethod(reservation *scCsv.ReservationData, ctx context.Context, tx *sql.Tx) (int, error) {
	if reservation.SalesAgentShopName != "" {
		id, err := getReservationMethod(reservation.SalesAgentShopName, ctx, tx)
		if err != nil {
			logging.Info(fmt.Sprintf("failed to get reservation method error: %v", err), nil)
		}
		if id == 0 {
			logging.Debug("start inserting reservation method master", nil)

			id, err = insertReservationMethod(reservation.SalesAgentShopName, ctx, tx)
			if err != nil {
				return 0, err
			}
			logging.Debug("finish inserting reservation method master", nil)

			return id, nil
		}
		return id, nil
	}
	return 0, fmt.Errorf("sales agent shop name is null")
}

func getReservationMethod(reservationMethod string, ctx context.Context, tx *sql.Tx) (int, error) {
	record, err := models.ReservationMethodMasters(
		qm.Where(models.ReservationMethodMasterColumns.ReservationMethodName+" = ?", reservationMethod),
	).One(ctx, tx)
	if err != nil {
		return 0, err
	}
	if record == nil {
		return 0, fmt.Errorf("%s not found", reservationMethod)
	}
	return record.ReservationMethodID, nil
}

func insertReservationMethod(reservationMethod string, ctx context.Context, tx *sql.Tx) (int, error) {
	newReservationMethodMaster := models.ReservationMethodMaster{
		//ReservationMethodID:   int,
		ReservationMethodName: null.StringFrom(reservationMethod),
	}
	if err := newReservationMethodMaster.Insert(ctx, tx, boil.Infer()); err != nil {
		return 0, err
	}
	return newReservationMethodMaster.ReservationMethodID, nil
}

func checkProductMaster(productId, productName string, ctx context.Context, tx *sql.Tx) *string {
	//	プランコード、プラン名がproduct_masterテーブルのproduct_id, product_nameに一致するかチェック
	if productId == "" || productName == "" {
		return nil
	}
	planId, err := getProductMaster(productId, productName, ctx, tx)
	if err != nil {
		logging.Info(fmt.Sprintf("failed to get product master error: %v", err), nil)
	}
	return planId
}

func getProductMaster(productId, productName string, ctx context.Context, tx *sql.Tx) (*string, error) {
	record, err := models.ProductMasters(
		qm.Where(models.ProductMasterColumns.ProductID+" = ?", productId),
		qm.And(models.ProductMasterColumns.ProductName+" = ?", productName),
	).One(ctx, tx)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, fmt.Errorf("%s not found", productId)
	}
	return &record.ProductID, nil
}

func validateReservationData(reservation *scCsv.ReservationData) []string {
	var validationStr []string
	if reservation.Name == "" {
		validationStr = append(validationStr, "団体名または代表者氏名 漢字")
	}
	if reservation.NameKana == "" {
		validationStr = append(validationStr, "団体名または代表者氏名(半角)")
	}
	if reservation.PhoneNumber == "" {
		validationStr = append(validationStr, "団体または代表者番号")
	}
	if reservation.PostalCode == "" {
		validationStr = append(validationStr, "団体または代表者郵便番号")
	}
	if reservation.HomeAddress == "" {
		validationStr = append(validationStr, "団体または代表者住所")
	}
	if reservation.ReservationHolder == "" {
		validationStr = append(validationStr, "予約者・会員名漢字")
	}
	if reservation.ReservationHolderKana == "" {
		validationStr = append(validationStr, "予約者・会員名カタカナ")
	}
	if reservation.StayDays == 0 {
		validationStr = append(validationStr, "泊数")
	}
	if reservation.NumberOfRooms == 0 {
		validationStr = append(validationStr, "利用客室合計数")
	}
	if reservation.NumberOfGuests == 0 {
		validationStr = append(validationStr, "お客様総合計人数")
	}
	if reservation.ProductName == "" {
		validationStr = append(validationStr, "プラン名")
	}
	return validationStr
}

func checkNewGuest(reservation *scCsv.ReservationData, ctx context.Context, tx *sql.Tx) *models.Guest {
	guestRecord, err := models.Guests(
		qm.Select(models.GuestColumns.GuestID),
		qm.Where(models.GuestColumns.Name+" = ?", reservation.Name),
		qm.And(models.GuestColumns.NameKana+" = ?", reservation.NameKana),
		qm.And(models.GuestColumns.PhoneNumber+" = ? or "+models.GuestColumns.PostalCode+" = ?", reservation.PhoneNumber, helper.PostalCodeFormat(reservation.PostalCode)),
	).One(ctx, tx)
	if err != nil {
		logging.Info(fmt.Sprintf("failed to check new guest error: %v", err), nil)
	}
	if guestRecord == nil {
		return nil
	}
	return guestRecord
}

func validateDeleteReservationData(reservation *scCsv.ReservationData) []string {
	var validationStr []string
	if reservation.Name == "" {
		validationStr = append(validationStr, "団体名または代表者氏名 漢字")
	}
	if reservation.NameKana == "" {
		validationStr = append(validationStr, "団体名または代表者氏名(半角)")
	}
	if reservation.PhoneNumber == "" {
		validationStr = append(validationStr, "団体または代表者番号")
	}
	return validationStr
}

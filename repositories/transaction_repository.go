package repositories

import (
	"database/sql"
	"fmt"
	"kasir-api/models"
)

type TransactionRepository interface {
	GetReport(startDate, endDate string) (*models.ReportResponse, error)
	CreateTransaction(items []models.CheckoutItem) (*models.Transaction, error)
}

type transactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (repo *transactionRepository) GetReport(startDate, endDate string) (*models.ReportResponse, error) {
	var response models.ReportResponse

	// 1. Total Revenue
	queryRevenue := "SELECT COALESCE(SUM(total_amount), 0) FROM transactions WHERE created_at BETWEEN $1 AND $2"
	err := repo.db.QueryRow(queryRevenue, startDate, endDate).Scan(&response.TotalRevenue)
	if err != nil {
		return nil, err
	}

	// 2. Total Transaction
	queryCount := "SELECT COUNT(*) FROM transactions WHERE created_at BETWEEN $1 AND $2"
	err = repo.db.QueryRow(queryCount, startDate, endDate).Scan(&response.TotalTransaction)
	if err != nil {
		return nil, err
	}

	// 3. Best Seller
	queryBestSeller := `
		SELECT p.name, SUM(td.quantity) as total_sold
		FROM transaction_details td
		JOIN transactions t ON td.transaction_id = t.id
		JOIN products p ON td.product_id = p.id
		WHERE t.created_at BETWEEN $1 AND $2
		GROUP BY p.name
		ORDER BY total_sold DESC
		LIMIT 1
	`
	err = repo.db.QueryRow(queryBestSeller, startDate, endDate).Scan(&response.BestSeller.Name, &response.BestSeller.TotalSold)
	if err != nil {
		if err == sql.ErrNoRows {
			// Handle case where no sales found
			response.BestSeller = models.BestSeller{Name: "", TotalSold: 0}
		} else {
			return nil, err
		}
	}

	return &response, nil
}

func (repo *transactionRepository) CreateTransaction(items []models.CheckoutItem) (*models.Transaction, error) {
	var (
		res *models.Transaction
	)

	tx, err := repo.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// inisialisasi subtotal -> jumlah total transaksi keseluruhan
	totalAmount := 0
	// inisialisasi modeling transactionDetails -> nanti kita insert ke db
	details := make([]models.TransactionDetail, 0)
	// loop setiap item
	for _, item := range items {
		var productName string
		var productID, price, stock int
		// get product dapet pricing
		err := tx.QueryRow("SELECT id, name, price, stock FROM products WHERE id=$1", item.ProductID).Scan(&productID, &productName, &price, &stock)
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product id %d not found", item.ProductID)
		}

		if err != nil {
			return nil, err
		}

		// hitung current total = quantity * pricing
		// ditambahin ke dalam subtotal
		subtotal := item.Quantity * price
		totalAmount += subtotal

		// kurangi jumlah stok
		_, err = tx.Exec("UPDATE products SET stock = stock - $1 WHERE id = $2", item.Quantity, productID)
		if err != nil {
			return nil, err
		}

		// item nya dimasukkin ke transactionDetails
		details = append(details, models.TransactionDetail{
			ProductID:   productID,
			ProductName: productName,
			Quantity:    item.Quantity,
			Subtotal:    subtotal,
		})
	}

	// insert transaction
	var transactionID int
	err = tx.QueryRow("INSERT INTO transactions (total_amount) VALUES ($1) RETURNING ID", totalAmount).Scan(&transactionID)
	if err != nil {
		return nil, err
	}

	// insert transaction details

	// custom bulk insert to avoid N+1
	if len(details) > 0 {
		query := "INSERT INTO transaction_details (transaction_id, product_id, quantity, subtotal) VALUES "
		vals := []interface{}{}
		for i, detail := range details {
			// Update the detail with transaction ID (struct logic)
			details[i].TransactionID = transactionID

			// construct placeholder ($1, $2, $3, $4), ($5, ...)
			n := i * 4
			query += fmt.Sprintf("($%d, $%d, $%d, $%d),", n+1, n+2, n+3, n+4)
			vals = append(vals, transactionID, detail.ProductID, detail.Quantity, detail.Subtotal)
		}
		// remove trailing comma
		query = query[:len(query)-1]

		// execute bulk insert
		_, err = tx.Exec(query, vals...)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	res = &models.Transaction{
		ID:          transactionID,
		TotalAmount: totalAmount,
		Details:     details,
	}

	return res, nil
}

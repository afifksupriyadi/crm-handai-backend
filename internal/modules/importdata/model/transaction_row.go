package model

// TransactionExcelRow represents one row from transaction Excel file
// File: Transaksi_Kasir_Warung_*.xlsx
type TransactionExcelRow struct {
	No               int
	NoStruk          string
	Tanggal          string
	Jam              string
	NamaPelanggan    string
	NamaProduk       string
	JumlahProduk     int
	HargaPerProduk   float64
	Varian           string
	HargaVarian      float64
	Subtotal         float64
	DiskonTransaksi  float64
	OngkosKirim      float64
	Total            float64
	Status           string
	MetodePembayaran string
}

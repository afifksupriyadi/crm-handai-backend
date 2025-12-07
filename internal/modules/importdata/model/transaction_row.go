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
	HargaPerProduk   int64
	Varian           string
	HargaVarian      int64
	Subtotal         int64
	DiskonTransaksi  int64
	OngkosKirim      int64
	Total            int64
	Status           string
	MetodePembayaran string
}

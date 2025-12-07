package model

// CustomerExcelRow represents one row from customer Excel file
// File: Transaksi_Pelanggan_Kasir_Warung_*.xlsx
type CustomerExcelRow struct {
	No            int
	NamaPelanggan string
	NomorTelepon  string
}

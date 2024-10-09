package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Struct untuk Menu Item
// Mewakili item menu dengan nama dan harga
type MenuItem struct {
	Name  string  // Nama item menu
	Price float64 // Harga item menu
}

// Struct untuk Pesanan
// Mewakili pesanan dengan daftar item dan total harga
type Order struct {
	MenuItems []MenuItem // Daftar item menu yang dipesan
	Total     float64    // Total harga dari pesanan
}

// Interface untuk manajemen menu
// Mendefinisikan metode yang harus diimplementasikan
type MenuManager interface {
	AddMenuItem(name string, price float64) // Menambahkan item menu
	PrintMenu()                             // Menampilkan daftar menu
}

// Struct Restaurant yang akan mengimplementasi interface MenuManager
type Restaurant struct {
	Menu []MenuItem // Daftar item menu yang tersedia
}

var wg sync.WaitGroup // WaitGroup untuk sinkronisasi goroutine

// Implementasi interface MenuManager
// Menambahkan item menu baru
func (r *Restaurant) AddMenuItem(name string, price float64) {
	r.Menu = append(r.Menu, MenuItem{Name: name, Price: price})
}

// Menampilkan daftar menu
func (r *Restaurant) PrintMenu() {
	fmt.Println("Menu:")
	for _, item := range r.Menu {
		fmt.Printf("%s: Rp%.2f\n", item.Name, item.Price)
	}
}

// Fungsi untuk menerima pesanan menggunakan goroutine dan channel
func takeOrder(restaurant *Restaurant, ch chan<- Order) {
	defer wg.Done() // Pastikan wg.Done dipanggil saat goroutine selesai
	order := Order{}
	var itemName string
	var itemQty int
	scanner := bufio.NewScanner(os.Stdin) // Scanner untuk membaca input pengguna

	for {
		// Menampilkan menu dan meminta nama item
		fmt.Println("Masukkan nama item (ketik 'selesai' untuk menyelesaikan): ")
		scanner.Scan()
		itemName = strings.ToLower(scanner.Text())

		if itemName == "selesai" {
			break // Jika pengguna mengetik 'selesai', keluar dari loop
		}

		// Validasi pesanan
		if menuItem, ok := validateOrderItem(restaurant, itemName); ok {
			fmt.Println("Masukkan jumlah: ")
			fmt.Scanln(&itemQty)
			order.MenuItems = append(order.MenuItems, *menuItem)
			order.Total += menuItem.Price * float64(itemQty) // Menghitung total harga
		} else {
			fmt.Println("Item tidak valid. Coba lagi.")
		}
	}
	// Kirim pesanan ke channel
	ch <- order
}

// Fungsi untuk memvalidasi item pesanan dari menu
func validateOrderItem(restaurant *Restaurant, itemName string) (*MenuItem, bool) {
	for _, menuItem := range restaurant.Menu {
		if strings.ToLower(menuItem.Name) == itemName {
			return &menuItem, true // Item ditemukan
		}
	}
	return nil, false // Item tidak valid
}

// Fungsi untuk memvalidasi input harga
func validatePrice(price string) (float64, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Terjadi kesalahan saat memvalidasi harga:", r)
		}
	}()
	matched, _ := regexp.MatchString(`^[0-9]+(\.[0-9]+)?$`, price) // Regex untuk validasi angka
	if !matched {
		return 0, fmt.Errorf("Format harga tidak valid") // Mengembalikan error jika format tidak valid
	}
	return strconv.ParseFloat(price, 64) // Mengonversi string ke float
}

// Fungsi untuk encode pesanan ke base64
func encodeOrder(order Order) string {
	orderDetails := ""
	for _, item := range order.MenuItems {
		orderDetails += fmt.Sprintf("%s:%.2f,", item.Name, item.Price) // Menyusun detail pesanan
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(orderDetails)) // Mengonversi ke base64
	return encoded
}

// Fungsi untuk menangani pembayaran
func handlePayment(totalOrder float64) {
	var priceInput string
	var price float64
	for {
		fmt.Println("Masukkan jumlah yang dibayar:")
		fmt.Scanln(&priceInput)

		// Validasi input pembayaran
		if validPrice, err := validatePrice(priceInput); err == nil {
			price = validPrice

			if price >= totalOrder {
				fmt.Printf("Jumlah yang dibayar valid. Kembalian: Rp%.2f\n", price-totalOrder)
				break
			} else {
				fmt.Println("Jumlah yang dibayar kurang dari total pesanan. Coba lagi.")
			}
		} else {
			fmt.Println("Input pembayaran tidak valid. Harap masukkan angka yang benar.")
		}
	}
}

func main() {
	restaurant := &Restaurant{}
	// Tambah menu menggunakan pointer dan method
	restaurant.AddMenuItem("Nasi Goreng", 25000)
	restaurant.AddMenuItem("Mie Goreng", 22000)
	restaurant.AddMenuItem("Ayam Bakar", 30000)
	// Menampilkan menu
	restaurant.PrintMenu()

	// Channel untuk pesanan
	orderChannel := make(chan Order)

	// Menggunakan goroutine untuk menerima pesanan
	wg.Add(1)
	go takeOrder(restaurant, orderChannel)

	// Tunggu semua goroutine selesai sebelum menutup channel
	go func() {
		wg.Wait()
		close(orderChannel) // Menutup channel setelah goroutine selesai
	}()

	var totalOrder float64

	// Mengambil pesanan dari channel
	for order := range orderChannel {
		fmt.Println("Pesanan Anda:")
		for _, item := range order.MenuItems {
			fmt.Printf("- %s\n", item.Name)
		}
		totalOrder += order.Total // Menghitung total keseluruhan pesanan
	}

	fmt.Printf("Total Pesanan: Rp%.2f\n", totalOrder)

	// Encode pesanan menggunakan base64
	encodedOrder := encodeOrder(Order{MenuItems: restaurant.Menu})
	fmt.Println("Pesanan (encoded base64):", encodedOrder)

	// Menangani pembayaran
	handlePayment(totalOrder)

	// Contoh penggunaan sync.WaitGroup untuk menunggu goroutine selesai
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("Memproses pesanan di goroutine lain...")
		time.Sleep(2 * time.Second) // Simulasi pemrosesan
	}()
	wg.Wait()

	fmt.Println("Program selesai")
}

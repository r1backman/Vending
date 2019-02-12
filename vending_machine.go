package main

import "fmt"

import (
  "net/http"
  "log"
  "html/template"
  "time"
  "database/sql"
  _ "github.com/go-sql-driver/mysql"
  "os"
  "strings"
  "strconv"
)

type VMPageVariables struct {
  Time             string
}

type OProductRow struct {
  Uid           int
  Description   string
  Amount        int
}
type OMoneyRow struct {
  Uid           int
  Description   string
  Amount        int
}

type OPageVariables struct {
  Time             string
  PageProductRow       []OProductRow
  PageMoneyRow         []OMoneyRow
}

type VProductRow struct {
  Description   string
  Amount        int
  Price         float32
}
type VMoneyRow struct {
  Description   string
  Amount        int
}

type VPageVariables struct {
  Time             string
  PageProductRow       []VProductRow
  PageMoneyRow         []VMoneyRow
  DispMessage      string
}

type CPageVariables struct {
  Time             string
}


func main() {
  fmt.Println("On port 8080, Ctl c to stop")
  http.Handle("/pics/", http.StripPrefix("/pics/", http.FileServer(http.Dir("pics"))))
  http.HandleFunc("/", DisplayVendingMachine)
  http.HandleFunc("/operator", DisplayOperator)
  http.HandleFunc("/vender", DisplayVending)
  http.HandleFunc("/client", DisplayClient)
  log.Fatal(http.ListenAndServe(":8080", nil))
}


func DisplayVendingMachine(w http.ResponseWriter, r *http.Request){
  now := time.Now()

  ClientPageVars := VMPageVariables{
    Time: now.Format("15:04:05"),
  }

  t, err := template.ParseFiles("html/vending_machine.html")
  if err != nil { // if there is an error
    log.Print("template parsing error: ", err) // log it
  }

  err = t.Execute(w, ClientPageVars) 
  if err != nil { // if there is an error
    log.Print("template executing error: ", err) //log it
  }

}

func checkErr(err error) {
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
}

func check_machine_money(db *sql.DB, key int, value int) string {
  var vendor_amt int

  err := db.QueryRow(`SELECT vendor_amt FROM money WHERE uid = ?;`, key).Scan(&vendor_amt)
  checkErr(err)
  if value > vendor_amt {
    return ("ERROR " + strconv.Itoa(value) + " > " + strconv.Itoa(vendor_amt))
  }
  return ("OK")
}

func take_machine_money(db *sql.DB, key int, value int) string {

  stmt, err := db.Prepare("update money set opr_amt = opr_amt + ?, vendor_amt = vendor_amt - ? where uid=?")
  checkErr(err)
  stmt.Exec(value, value, key)
  checkErr(err)

  return ("OK")
}

func check_machine_product_money(db *sql.DB, table string, key int, value int) string {
  var opr_amt int

  if table == "p" {
    err := db.QueryRow(`SELECT opr_amt FROM product WHERE uid = ?;`, key).Scan(&opr_amt)
    checkErr(err)
  } else {
    err := db.QueryRow(`SELECT opr_amt FROM money WHERE uid = ?;`, key).Scan(&opr_amt)
    checkErr(err)
  }
  //fmt.Println(table, key, value, opr_amt)
  if value > opr_amt {
    return ("ERROR " + strconv.Itoa(value) + " > " + strconv.Itoa(opr_amt))
  }
  return ("OK")
}

func load_machine(db *sql.DB, table string, key int, value int) string {

  if table == "p" {
    stmt, err := db.Prepare("update product set opr_amt = opr_amt - ?, vendor_amt = vendor_amt + ? where uid=?")
    checkErr(err)
    stmt.Exec(value, value, key)
    checkErr(err)
  } else {
    stmt, err := db.Prepare("update money set opr_amt = opr_amt - ?, vendor_amt = vendor_amt + ? where uid=?")
    checkErr(err)
    stmt.Exec(value, value, key)
    checkErr(err)
  }

  return ("OK")
}

func DisplayOperator(w http.ResponseWriter, r *http.Request){

  var message = "OK"

  db, err := sql.Open("mysql", "roger:abcdefgh@tcp(127.0.0.1:3306)/vendor?charset=utf8")
  checkErr(err)

  r.ParseForm()
  
  takestuff := r.Form.Get("take")
  loadstuff := r.Form.Get("load")
  
  if takestuff == "take money" {
    for k, v := range r.Form {
      table := k[0:1]
      value := strings.Join(v, "")
      if (value != "take money" && table == "r") {
        key := k[1:]
        n_key, err := strconv.Atoi(key)
        checkErr(err)
        n_value, err := strconv.Atoi(value)
        checkErr(err)
        message = check_machine_money(db, n_key, n_value)
      }
      if message != "OK" {
        break
      }
    }
    if message == "OK" {
      for k, v := range r.Form {
        table := k[0:1]
        value := strings.Join(v, "")
        if (value != "take money" && table == "r") {
          key := k[1:]
          n_key, err := strconv.Atoi(key)
          checkErr(err)
          n_value, err := strconv.Atoi(value)
          checkErr(err)
          message = take_machine_money(db, n_key, n_value)
        }
        if message != "OK" {
          break
        }
      }
    }
  }
  if loadstuff == "load" {
    for k, v := range r.Form {
      table := k[0:1]
      value := strings.Join(v, "")
      if (value != "load") {
        key := k[1:]
        n_key, err := strconv.Atoi(key)
        checkErr(err)
        n_value, err := strconv.Atoi(value)
        checkErr(err)
        message = check_machine_product_money(db, table, n_key, n_value)
      }
      if message != "OK" {
        break
      }
    }
    if message == "OK" {
      for k, v := range r.Form {
        table := k[0:1]
        value := strings.Join(v, "")
        if (value != "load") {
          key := k[1:]
          n_key, err := strconv.Atoi(key)
          checkErr(err)
          n_value, err := strconv.Atoi(value)
          checkErr(err)
          message = load_machine(db, table, n_key, n_value)
        }
        if message != "OK" {
          break
        }
      }
    }
  }

  stmt, err := db.Prepare("update message set message_text=? where uid=?")
  checkErr(err)
  stmt.Exec(message, 1)
  checkErr(err)

  now := time.Now()

  product_rows, err := db.Query("SELECT uid, p_description, opr_amt FROM product")
  checkErr(err)
  money_rows, err := db.Query("SELECT uid, r_description, opr_amt FROM money")
  checkErr(err)
  
  var MyProductRow []OProductRow
  var MyMoneyRow   []OMoneyRow

  for product_rows.Next() {
    var uid     int
    var product string
    var opr_amt int

    err = product_rows.Scan(&uid, &product, &opr_amt)
    checkErr(err)
    MyProductRow = append(MyProductRow, OProductRow{uid, product, opr_amt})
  }

  for money_rows.Next() {
    var uid     int
    var money   string
    var opr_amt int

    err = money_rows.Scan(&uid, &money, &opr_amt)
    checkErr(err)
    MyMoneyRow = append(MyMoneyRow, OMoneyRow{uid, money, opr_amt})
  }
  
  OperatorPageVars := OPageVariables{
    Time: now.Format("15:04:05"),
    PageProductRow : MyProductRow,
    PageMoneyRow   : MyMoneyRow,
  }

  t, err := template.ParseFiles("html/operator.html")
  if err != nil { // if there is an error
    log.Print("template parsing error: ", err) // log it
  }

  err = t.Execute(w, OperatorPageVars) 
  if err != nil { // if there is an error
    log.Print("template executing error: ", err) //log it
  }
  db.Close()
}

func DisplayVending(w http.ResponseWriter, r *http.Request){
  now := time.Now()

  db, err := sql.Open("mysql", "roger:abcdefgh@tcp(127.0.0.1:3306)/vendor?charset=utf8")
  checkErr(err)

  product_rows, err := db.Query("SELECT p_description, vendor_amt, price FROM product")
  checkErr(err)
  money_rows, err := db.Query("SELECT r_description, vendor_amt FROM money")
  checkErr(err)
  
  type Tag struct {
    ID   string    `json:"message_text"`
  }
  var tag Tag

  err = db.QueryRow("SELECT message_text FROM message WHERE uid = 1").Scan(&tag.ID)
  checkErr(err)

  var MyProductRow []VProductRow
  var MyMoneyRow   []VMoneyRow

  for product_rows.Next() {
    var product string
    var vendor_amt int
    var price float32

    err = product_rows.Scan(&product, &vendor_amt, &price)
    checkErr(err)
    MyProductRow = append(MyProductRow, VProductRow{product, vendor_amt, price})
  }

  for money_rows.Next() {
    var money string
    var opr_amt int

    err = money_rows.Scan(&money, &opr_amt)
    checkErr(err)
    MyMoneyRow = append(MyMoneyRow, VMoneyRow{money, opr_amt})
  }

  OperatorPageVars := VPageVariables{
    Time: now.Format("15:04:05"),
    PageProductRow : MyProductRow,
    PageMoneyRow   : MyMoneyRow,
    DispMessage    : tag.ID,
  }

  t, err := template.ParseFiles("html/vending.html")
  if err != nil { // if there is an error
    log.Print("template parsing error: ", err) // log it
  }

  err = t.Execute(w, OperatorPageVars) 
  if err != nil { // if there is an error
    log.Print("template executing error: ", err) //log it
  }
  db.Close()
}

func check_client(db *sql.DB, table string, key int, value int) string {
  var client_amt int
  var vendor_amt int

  if table == "p" {
    err := db.QueryRow(`SELECT vendor_amt FROM product WHERE uid = ?;`, key).Scan(&vendor_amt)
    checkErr(err)
    if value > vendor_amt {
      return ("ERROR PRODUCT " + strconv.Itoa(value) + " > " + strconv.Itoa(vendor_amt))
    }
  } else {
    err := db.QueryRow(`SELECT client_amt FROM money WHERE uid = ?;`, key).Scan(&client_amt)
    checkErr(err)
    if value > client_amt {
      return ("ERROR MONEY " + strconv.Itoa(value) + " > " + strconv.Itoa(client_amt))
    }
  }
  return ("OK")
}

func get_product_price(db *sql.DB, key int) float32 {
  var price float32

  err := db.QueryRow(`SELECT price FROM product WHERE uid = ?;`, key).Scan(&price)
  checkErr(err)
  return (price)
}

func get_rand_amt(db *sql.DB, key int) float32 {
  var r_description string
  var price float32

  err := db.QueryRow(`SELECT r_description FROM money WHERE uid = ?;`, key).Scan(&r_description)
  checkErr(err)
  switch r_description {
    case "R200":
      price = 200
    case "R100":
      price = 100
    case "R50":
      price = 50
    case "R20":
      price = 20
    case "R10":
      price = 10
    case "R1":
      price = 1
    case "R.50":
      price = 0.5
    case "R.20":
      price = 0.2
    case "R.10":
      price = 0.1
    case "R.05":
      price = 0.05
  }
  return (price)
}

func DisplayClient(w http.ResponseWriter, r *http.Request){

  var message = "OK"

  db, err := sql.Open("mysql", "roger:abcdefgh@tcp(127.0.0.1:3306)/vendor?charset=utf8")
  checkErr(err)

  r.ParseForm()
  
  loadstuff := r.Form.Get("buy")
  
  if loadstuff == "buy" {
    var total_items int = 0
    var total_amt_to_buy float32 = 0
    var total_amt_tendered float32 = 0

    type MoneyChange struct {
      Uid           int
      Description   string
      rand_amt      float32
      vendor_amt    int
      client_amt    int
      pay_amt       int
    }
    var MyChange []MoneyChange

    for k, v := range r.Form {
      table := k[0:1]
      value := strings.Join(v, "")
      if (value != "buy") {
        key := k[1:]
        n_key, err := strconv.Atoi(key)
        checkErr(err)
        n_value, err := strconv.Atoi(value)
        checkErr(err)
        message = check_client(db, table, n_key, n_value)
      }
      if message != "OK" {
        break
      }
    }
    if message == "OK" {
      for k, v := range r.Form {
        table := k[0:1]
        value := strings.Join(v, "")
        if (value != "buy" && table == "p") {
          key := k[1:]
          n_key, err := strconv.Atoi(key)
          checkErr(err)
          n_value, err := strconv.Atoi(value)
          checkErr(err)
          if n_value > 0 {
            total_items += n_value
            total_amt_to_buy += (float32(n_value) * get_product_price(db, n_key))
          }
        }
        if (value != "buy" && table == "r") {
          key := k[1:]
          n_key, err := strconv.Atoi(key)
          checkErr(err)
          n_value, err := strconv.Atoi(value)
          checkErr(err)
          if n_value > 0 {
            total_amt_tendered += (float32(n_value) * get_rand_amt(db, n_key))
          }
        }
      }
      if total_items == 0 {
        message = "ERROR NO ITEMS SELECTED"
      } else if total_amt_tendered == 0 {
        message = "ERROR NO MONEY"
      } else if total_amt_tendered < total_amt_to_buy {
        message = "ERROR NOT ENOUGH MONEY"
      }
      //fmt.Println(total_amt_to_buy)
    }
    if message == "OK" {
      money_rows, err := db.Query(`SELECT uid, r_description, vendor_amt, client_amt FROM money order by uid;`)
      checkErr(err)
      for money_rows.Next() {
        var r_description   string
        var vendor_amt int
        var client_amt int
        var rand_amt float32
        var uid      int

        err = money_rows.Scan(&uid, &r_description, &vendor_amt, &client_amt)
        checkErr(err)
        switch r_description {
        case "R200":
          rand_amt = 200
        case "R100":
          rand_amt = 100
        case "R50":
          rand_amt = 50
        case "R20":
          rand_amt = 20
        case "R10":
          rand_amt = 10
        case "R1":
          rand_amt = 1
        case "R.50":
          rand_amt = 0.5
        case "R.20":
          rand_amt = 0.2
        case "R.10":
          rand_amt = 0.1
        case "R.05":
          rand_amt = 0.05
        }
        MyChange = append(MyChange, MoneyChange{uid, r_description, rand_amt, vendor_amt, client_amt, 0})
      }
      for k, v := range r.Form {
        table := k[0:1]
        value := strings.Join(v, "")
        if (value != "buy" && table == "r") {
          key := k[1:]
          n_key, err := strconv.Atoi(key)
          checkErr(err)
          n_value, err := strconv.Atoi(value)
          checkErr(err)
          for index, element_t := range MyChange {
            if element_t.Uid == n_key {
              //fmt.Println("Found", n_value, n_key)
              MyChange[index].pay_amt = n_value
              break
            }
          }
        }
      }
      
      for index, element_t := range MyChange {
        if element_t.pay_amt > 0 {
          MyChange[index].vendor_amt += MyChange[index].pay_amt
          MyChange[index].client_amt -= MyChange[index].pay_amt
        }
      }
      
      var change_client float32 = total_amt_tendered - total_amt_to_buy

      for index, element_t := range MyChange {
        for element_t.rand_amt <= change_client {
          if MyChange[index].vendor_amt > 0 {
            change_client -= MyChange[index].rand_amt
            MyChange[index].vendor_amt -= 1
            MyChange[index].client_amt += 1
          } else {
            break
          }
        }
      }

      if change_client != 0 {
        message = "ERROR NOT ENOUGH CHANGE"
      }
    }
    //update db
    if message == "OK" {
      //fmt.Println(MyChange)
      //fmt.Println("Update")
      for index, element_t := range MyChange {
        stmt, err := db.Prepare("update money set vendor_amt=?, client_amt=? where uid=?")
        stmt.Exec(MyChange[index].vendor_amt, MyChange[index].client_amt, element_t.Uid)
        checkErr(err)
        //fmt.Println(index, element_t.Uid, MyChange[index].vendor_amt, MyChange[index].client_amt)
      }
      for k, v := range r.Form {
        table := k[0:1]
        value := strings.Join(v, "")
        if (value != "buy" && table == "p") {
          key := k[1:]
          n_key, err := strconv.Atoi(key)
          checkErr(err)
          n_value, err := strconv.Atoi(value)
          checkErr(err)
          if n_value > 0 {
            stmt, err := db.Prepare("update product set vendor_amt = vendor_amt - ?, client_amt = client_amt + ? where uid=?")
            stmt.Exec(n_value, n_value, n_key)
            checkErr(err)
          }
        }
      }
    }
  }

  stmt, err := db.Prepare("update message set message_text=? where uid=?")
  checkErr(err)
  stmt.Exec(message, 1)
  checkErr(err)

  now := time.Now()

  product_rows, err := db.Query("SELECT uid, p_description, client_amt FROM product")
  checkErr(err)
  money_rows, err := db.Query("SELECT uid, r_description, client_amt FROM money")
  checkErr(err)
  
  var MyProductRow []OProductRow
  var MyMoneyRow   []OMoneyRow

  for product_rows.Next() {
    var uid     int
    var product string
    var client_amt int

    err = product_rows.Scan(&uid, &product, &client_amt)
    checkErr(err)
    MyProductRow = append(MyProductRow, OProductRow{uid, product, client_amt})
  }

  for money_rows.Next() {
    var uid     int
    var money   string
    var client_amt int

    err = money_rows.Scan(&uid, &money, &client_amt)
    checkErr(err)
    MyMoneyRow = append(MyMoneyRow, OMoneyRow{uid, money, client_amt})
  }
  
  OperatorPageVars := OPageVariables{
    Time: now.Format("15:04:05"),
    PageProductRow : MyProductRow,
    PageMoneyRow   : MyMoneyRow,
  }

  t, err := template.ParseFiles("html/client.html")
  if err != nil { // if there is an error
    log.Print("template parsing error: ", err) // log it
  }

  err = t.Execute(w, OperatorPageVars) 
  if err != nil { // if there is an error
    log.Print("template executing error: ", err) //log it
  }
  db.Close()
}

import sqlite3
import gspread
from oauth2client.service_account import ServiceAccountCredentials

# Google Sheets API Setup
SHEET_ID = "1TTN-grLVzcjQjebJtX5VQ4gUfC-uxMjym2-kZNoB3so"  # Replace with your actual Google Sheet ID
RANGE_NAME = "A:D"  # Only use column references

# Load Google Sheets API credentials
creds = ServiceAccountCredentials.from_json_keyfile_name("service-account.json", ["https://spreadsheets.google.com/feeds", "https://www.googleapis.com/auth/drive"])
client = gspread.authorize(creds)
sheet = client.open_by_key(SHEET_ID).worksheet("Sheet2")  # Ensure you're opening "Sheet2"

# Connect to SQLite Database
conn = sqlite3.connect("prod.sqlite")
cursor = conn.cursor()

# Query to get users who have paid (i.e., transactions exist) but haven't been pushed yet
query = """
SELECT u.id, u.name, t.id AS txn_id, t.amount 
FROM users u
JOIN transactions t ON u.id = t.user_id
LEFT JOIN pushed_txn p ON t.id = p.txn_id
WHERE t.is_verified = FALSE AND p.txn_id IS NULL;
"""

cursor.execute(query)
data = cursor.fetchall()

if data:
    # Format data for Google Sheets
    rows = [["User ID", "User Name", "Transaction ID", "Amount"]]  # Headers
    rows.extend(data)

    # Push data to Google Sheet
    sheet.append_rows(rows, value_input_option="RAW")  # Appends without clearing

    # Insert pushed transactions into pushed_txn table
    txn_ids = [(txn[2],) for txn in data]  # Extract only txn_id
    cursor.executemany("INSERT INTO pushed_txn (txn_id) VALUES (?)", txn_ids)
    conn.commit()

    print(f"Successfully pushed {len(data)} new transactions to Google Sheets.")
else:
    print("No new transactions to push.")

# Close DB Connection
conn.close()

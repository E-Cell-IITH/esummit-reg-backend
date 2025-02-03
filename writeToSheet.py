import sqlite3
import gspread
from oauth2client.service_account import ServiceAccountCredentials

# Google Sheets API Setup
SHEET_ID = "1TTN-grLVzcjQjebJtX5VQ4gUfC-uxMjym2-kZNoB3so"  # Replace with your actual Google Sheet ID
RANGE_NAME = "Sheet1!A:D"  # Adjust based on your sheet

# Load Google Sheets API credentials
creds = ServiceAccountCredentials.from_json_keyfile_name("service-account.json", ["https://spreadsheets.google.com/feeds", "https://www.googleapis.com/auth/drive"])
client = gspread.authorize(creds)
sheet = client.open_by_key(SHEET_ID).worksheet("Sheet2")

# Connect to SQLite Database
conn = sqlite3.connect("prod.sqlite")
cursor = conn.cursor()

# Query to get users who have paid (i.e., transactions exist)
query = """
SELECT u.id, u.name, t.id AS txn_id, t.amount 
FROM users u
JOIN transactions t ON u.id = t.user_id
WHERE t.is_verified = TRUE;
"""

cursor.execute(query)
data = cursor.fetchall()

# Format data for Google Sheets
rows = [["User ID", "User Name", "Transaction ID", "Amount"]]  # Headers
rows.extend(data)

# Push data to Google Sheet
sheet.clear()  # Optional: Clears the sheet before writing new data
sheet.update(RANGE_NAME, rows)

# Close DB Connection
conn.close()

print("Data successfully pushed to Google Sheets!")

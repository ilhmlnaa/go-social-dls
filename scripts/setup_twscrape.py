#!/usr/bin/env python3
"""
Setup script for twscrape
Helps add Twitter accounts for scraping
"""
import asyncio
import sys
from twscrape import API

async def add_account(username, password, email, email_password):
    """Add a Twitter account to twscrape"""
    api = API()
    
    try:
        await api.pool.add_account(username, password, email, email_password)
        print(f"✅ Account {username} added successfully!")
        print("⏳ Logging in to verify...")
        
        await api.pool.login_all()
        print("✅ Login successful!")
        
        # Show account info
        accounts = await api.pool.accounts_info()
        print(f"\n📊 Total accounts configured: {len(accounts)}")
        for acc in accounts:
            print(f"   - {acc.username}: {acc.status}")
            
    except Exception as e:
        print(f"❌ Error: {e}")
        sys.exit(1)

async def list_accounts():
    """List all configured accounts"""
    api = API()
    accounts = await api.pool.accounts_info()
    
    if not accounts:
        print("⚠️  No accounts configured yet.")
        print("\nTo add an account, run:")
        print("  python3 scripts/setup_twscrape.py add <username> <password> <email> <email_password>")
        return
    
    print(f"📊 Total accounts: {len(accounts)}\n")
    for acc in accounts:
        print(f"Username: {acc.username}")
        print(f"Status: {acc.status}")
        print(f"Locks: {acc.locks}")
        print("-" * 50)

def main():
    if len(sys.argv) < 2:
        print("Usage:")
        print("  Add account: python3 setup_twscrape.py add <username> <password> <email> <email_password>")
        print("  List accounts: python3 setup_twscrape.py list")
        sys.exit(1)
    
    command = sys.argv[1]
    
    if command == "add":
        if len(sys.argv) < 6:
            print("Error: Missing arguments")
            print("Usage: python3 setup_twscrape.py add <username> <password> <email> <email_password>")
            sys.exit(1)
        
        username = sys.argv[2]
        password = sys.argv[3]
        email = sys.argv[4]
        email_password = sys.argv[5]
        
        asyncio.run(add_account(username, password, email, email_password))
    
    elif command == "list":
        asyncio.run(list_accounts())
    
    else:
        print(f"Unknown command: {command}")
        sys.exit(1)

if __name__ == "__main__":
    main()

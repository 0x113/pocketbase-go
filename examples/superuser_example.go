package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/0x113/pocketbase-go"
)

// SuperuserExample demonstrates superuser authentication and user impersonation
func SuperuserExample() {
	fmt.Println("=== Superuser & Impersonation Example ===")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create a separate client for superuser operations
	superuserClient := CreateSuperuserClient("http://localhost:8090")

	// Authenticate as superuser
	// Replace with actual superuser credentials from your PocketBase instance
	superuser, err := superuserClient.AuthenticateAsSuperuser(ctx, "admin@example.com", "admin_password")
	if err != nil {
		fmt.Printf("Superuser authentication failed: %v\n", err)
		fmt.Println("This is expected if you don't have a superuser configured")
		fmt.Println("Skipping impersonation demo...")
		return
	}

	fmt.Printf("[SUCCESS] Authenticated as superuser: %v\n", superuser["email"])
	fmt.Printf("[SUCCESS] Superuser ID: %v\n", superuser["id"])
	fmt.Printf("[SUCCESS] Superuser Token: %s\n", superuserClient.GetToken())
	fmt.Println()

	// Get some user records to impersonate
	fmt.Println("Looking for users to impersonate...")
	userRecords, err := superuserClient.GetAllRecords(ctx, "users")
	if err != nil {
		log.Printf("Failed to fetch users for impersonation: %v", err)
		return
	}

	if len(userRecords) == 0 {
		fmt.Println("No user records available to demonstrate impersonation")
		return
	}

	// Impersonate the first user
	userID := fmt.Sprintf("%v", userRecords[0]["id"])
	fmt.Printf("Impersonating user: %s\n", userID)

	// Impersonate user for 30 minutes (1800 seconds)
	impersonateResult, err := superuserClient.Impersonate(ctx, "users", userID, 1800,
		pocketbase.WithExpand("profile"),
		pocketbase.WithFields("id", "email", "username", "profile"))
	if err != nil {
		fmt.Printf("[ERROR] Impersonation failed: %v\n", err)
	} else {
		fmt.Printf("[SUCCESS] Successfully impersonated user: %v\n", impersonateResult.Record["email"])
		fmt.Printf("[SUCCESS] Impersonation token: %.50s...\n", impersonateResult.Token)

		// Create a new client with the impersonation token
		impersonatedClient := CreateClient("http://localhost:8090")
		impersonatedClient.SetToken(impersonateResult.Token)

		// Now make requests as the impersonated user
		fmt.Println("\nMaking requests as impersonated user...")
		userDataRecords, err := impersonatedClient.GetAllRecords(ctx, "user_data")
		if err != nil {
			fmt.Printf("[ERROR] Request as impersonated user failed: %v\n", err)
		} else {
			fmt.Printf("[SUCCESS] Fetched %d records as impersonated user\n", len(userDataRecords))
			if len(userDataRecords) > 0 {
				fmt.Printf("[SUCCESS] Sample data: %v\n", userDataRecords[0])
			}
		}
	}

	fmt.Println()
}

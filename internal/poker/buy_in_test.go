package poker

import (
	"testing"
	"time"
)

// TestBuyInConfiguration prueba la configuración de buy-in
func TestBuyInConfiguration(t *testing.T) {
	engine := NewPokerEngine()
	
	// Crear mesa con configuración personalizada
	config := TableConfig{
		SmallBlind:   5,
		BigBlind:     10,
		BuyInAmount:  500,
		MinBuyIn:     200,
		MaxBuyIn:     1000,
		IsCashGame:   true,
		AutoRestart:  true,
		RestartDelay: 3 * time.Second,
	}
	
	table := engine.CreateTableWithConfig("buy_in_test", config)
	
	// Verificar configuración
	if table.SmallBlind != 5 {
		t.Errorf("Expected SmallBlind 5, got %d", table.SmallBlind)
	}
	if table.BigBlind != 10 {
		t.Errorf("Expected BigBlind 10, got %d", table.BigBlind)
	}
	if table.BuyInAmount != 500 {
		t.Errorf("Expected BuyInAmount 500, got %d", table.BuyInAmount)
	}
	if table.MinBuyIn != 200 {
		t.Errorf("Expected MinBuyIn 200, got %d", table.MinBuyIn)
	}
	if table.MaxBuyIn != 1000 {
		t.Errorf("Expected MaxBuyIn 1000, got %d", table.MaxBuyIn)
	}
	if !table.IsCashGame {
		t.Errorf("Expected IsCashGame to be true")
	}
}

// TestAddPlayerWithBuyIn prueba agregar jugadores con buy-in personalizado
func TestAddPlayerWithBuyIn(t *testing.T) {
	engine := NewPokerEngine()
	table := engine.CreateTable("buy_in_player_test")
	
	// Agregar Alice con buy-in de 750
	_, err := engine.AddPlayerWithBuyIn("buy_in_player_test", "alice_id", "Alice", 750)
	if err != nil {
		t.Fatalf("Error adding Alice: %v", err)
	}
	
	// Agregar Bob con buy-in de 1200
	_, err = engine.AddPlayerWithBuyIn("buy_in_player_test", "bob_id", "Bob", 1200)
	if err != nil {
		t.Fatalf("Error adding Bob: %v", err)
	}
	
	// Verificar stacks
	if table.Players[0].Stack != 750 {
		t.Errorf("Expected Alice stack 750, got %d", table.Players[0].Stack)
	}
	if table.Players[1].Stack != 1200 {
		t.Errorf("Expected Bob stack 1200, got %d", table.Players[1].Stack)
	}
}

// TestBuyInValidation prueba la validación de buy-in
func TestBuyInValidation(t *testing.T) {
	engine := NewPokerEngine()
	engine.CreateTable("validation_test")
	
	// Intentar buy-in menor que el mínimo
	_, err := engine.AddPlayerWithBuyIn("validation_test", "alice_id", "Alice", 300)
	if err == nil {
		t.Errorf("Expected error for buy-in below minimum")
	}
	
	// Intentar buy-in mayor que el máximo
	_, err = engine.AddPlayerWithBuyIn("validation_test", "bob_id", "Bob", 2500)
	if err == nil {
		t.Errorf("Expected error for buy-in above maximum")
	}
	
	// Buy-in válido
	_, err = engine.AddPlayerWithBuyIn("validation_test", "carol_id", "Carol", 800)
	if err != nil {
		t.Errorf("Unexpected error for valid buy-in: %v", err)
	}
}

// TestGetTableConfig prueba obtener configuración de mesa
func TestGetTableConfig(t *testing.T) {
	engine := NewPokerEngine()
	engine.CreateTable("config_test")
	
	config, err := engine.GetTableConfig("config_test")
	if err != nil {
		t.Fatalf("Error getting table config: %v", err)
	}
	
	// Verificar valores por defecto
	if config.SmallBlind != 10 {
		t.Errorf("Expected default SmallBlind 10, got %d", config.SmallBlind)
	}
	if config.BigBlind != 20 {
		t.Errorf("Expected default BigBlind 20, got %d", config.BigBlind)
	}
	if config.BuyInAmount != 1000 {
		t.Errorf("Expected default BuyInAmount 1000, got %d", config.BuyInAmount)
	}
}

// TestUpdateTableConfig prueba actualizar configuración de mesa
func TestUpdateTableConfig(t *testing.T) {
	engine := NewPokerEngine()
	engine.CreateTable("update_config_test")
	
	newConfig := TableConfig{
		SmallBlind:   25,
		BigBlind:     50,
		BuyInAmount:  2000,
		MinBuyIn:     1000,
		MaxBuyIn:     5000,
		IsCashGame:   false,
		AutoRestart:  false,
		RestartDelay: 10 * time.Second,
	}
	
	err := engine.UpdateTableConfig("update_config_test", newConfig)
	if err != nil {
		t.Fatalf("Error updating table config: %v", err)
	}
	
	// Verificar que se aplicó la configuración
	config, err := engine.GetTableConfig("update_config_test")
	if err != nil {
		t.Fatalf("Error getting updated config: %v", err)
	}
	
	if config.SmallBlind != 25 {
		t.Errorf("Expected SmallBlind 25, got %d", config.SmallBlind)
	}
	if config.BigBlind != 50 {
		t.Errorf("Expected BigBlind 50, got %d", config.BigBlind)
	}
	if config.BuyInAmount != 2000 {
		t.Errorf("Expected BuyInAmount 2000, got %d", config.BuyInAmount)
	}
	if config.IsCashGame {
		t.Errorf("Expected IsCashGame to be false")
	}
}

// TestValidateBuyIn prueba la función de validación independiente
func TestValidateBuyIn(t *testing.T) {
	engine := NewPokerEngine()
	engine.CreateTable("validate_test")
	
	// Validar buy-in válido
	err := engine.ValidateBuyIn("validate_test", 800)
	if err != nil {
		t.Errorf("Unexpected error for valid buy-in: %v", err)
	}
	
	// Validar buy-in inválido (muy bajo)
	err = engine.ValidateBuyIn("validate_test", 300)
	if err == nil {
		t.Errorf("Expected error for buy-in below minimum")
	}
	
	// Validar buy-in inválido (muy alto)
	err = engine.ValidateBuyIn("validate_test", 2500)
	if err == nil {
		t.Errorf("Expected error for buy-in above maximum")
	}
}
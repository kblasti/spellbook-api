package api

import (
	"net/http"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/kblasti/spellbook/internal/database"
	"github.com/kblasti/spellbook/internal/auth"
)

type Character struct{
	ID			uuid.UUID		`json:"id"`
	Name		string			`json:"name"`
	ClassLevels	json.RawMessage	`json:"class_levels"`
}

func (cfg *APIConfig) HandlerCreateCharacter(w http.ResponseWriter, r *http.Request) {
	type Input struct{
		Name		string			`json:"name"`
		ClassLevels	json.RawMessage	`json:"class_levels"`
	}

	token, err := auth.GetBearerToken(r.Header)
    if err != nil {
        respondWithError(w, 401, "Error retrieving token")
        return
    }

    userID, _, err := auth.ValidateJWT(token, cfg.Secret)
    if err != nil {
        respondWithError(w, 401, "Error validating token")
        return
    }

	decoder := json.NewDecoder(r.Body)
    input := Input{}

    err = decoder.Decode(&input)
    if err != nil {
        respondWithError(w, 500, "Error decoding input")
        return
    }

	character, err := cfg.DB.CreateCharacter(r.Context(), database.CreateCharacterParams{
		Name:			input.Name,
		ClassLevels:	input.ClassLevels,
		UserID:			userID,
	})
	if err != nil {
		respondWithError(w, 500, "Error adding character to database")
		return
	}

	val := Character{
		ID:				character.ID,
		Name:			character.Name,
		ClassLevels:	character.ClassLevels,
	}

	respondWithJSON(w, 201, val)
	return
}

func (cfg *APIConfig) HandlerUpdateCharacter(w http.ResponseWriter, r *http.Request) {
	type Input struct{
		ID			uuid.UUID		`json:"id"`
		Name		string			`json:"name"`
		ClassLevels	json.RawMessage	`json:"class_levels"`
	}

	token, err := auth.GetBearerToken(r.Header)
    if err != nil {
        respondWithError(w, 401, "Error retrieving token")
        return
    }

    _, _, err = auth.ValidateJWT(token, cfg.Secret)
    if err != nil {
        respondWithError(w, 401, "Error validating token")
        return
    }

	decoder := json.NewDecoder(r.Body)
    input := Input{}

    err = decoder.Decode(&input)
    if err != nil {
        respondWithError(w, 500, "Error decoding input")
        return
    }

	character, err := cfg.DB.UpdateCharacter(r.Context(), database.UpdateCharacterParams{
		Name:			input.Name,
		ClassLevels:	input.ClassLevels,
		ID:				input.ID,
	})
	if err != nil {
		respondWithError(w, 500, "Error updating character")
		return
	}

	val := Character{
		ID:		character.ID,
		Name:			character.Name,
		ClassLevels:	character.ClassLevels,
	}

	respondWithJSON(w, 201, val)
	return
}

func (cfg *APIConfig) HandlerDeleteCharacter(w http.ResponseWriter, r *http.Request) {
	type Input struct {
		ID uuid.UUID `json:"id"`
	}

	token, err := auth.GetBearerToken(r.Header)
    if err != nil {
        respondWithError(w, 401, "Error retrieving token")
        return
    }

    userID, _, err := auth.ValidateJWT(token, cfg.Secret)
    if err != nil {
        respondWithError(w, 401, "Error validating token")
        return
    }

	decoder := json.NewDecoder(r.Body)
    input := Input{}

    err = decoder.Decode(&input)
    if err != nil {
        respondWithError(w, 500, "Error decoding input")
        return
    }

	err = cfg.DB.DeleteCharacter(r.Context(), database.DeleteCharacterParams{
		ID:     input.ID,
		UserID: userID,
	})

	respondWithMessage(w, 201, "Character deleted")
	return
}

func (cfg *APIConfig) HandlerGetSpellSlots(w http.ResponseWriter, r *http.Request) { 
	type Input struct { 
		ID		uuid.UUID	`json:"id"`
		Name 	string 		`json:"name"` 
	} 
	
	type SpellSlotsResponse struct { 
		FullCasterSlots json.RawMessage `json:"full_caster_slots"` 
		WarlockSlots json.RawMessage `json:"warlock_slots"` 
	} 
	
	token, err := auth.GetBearerToken(r.Header) 
	if err != nil { 
		respondWithError(w, 401, "Error retrieving token") 
		return 
	} 
	
	_, _, err = auth.ValidateJWT(token, cfg.Secret) 
	if err != nil { 
		respondWithError(w, 401, "Error validating token") 
		return 
	} 
	
	decoder := json.NewDecoder(r.Body)
    input := Input{}

    err = decoder.Decode(&input)
    if err != nil {
        respondWithError(w, 500, "Error decoding input")
        return
    }
	
	classLevelsJSON, err := cfg.DB.GetClassLevels(r.Context(), input.ID) 
	if err != nil { 
		respondWithError(w, 500, "Error getting class levels") 
		return 
	} 
	
	var levels map[string]int 
	err = json.Unmarshal(classLevelsJSON, &levels) 
	if err != nil { 
		respondWithError(w, 500, "Error parsing class levels") 
		return 
	} 
	
	warlockLevel := levels["warlock"] 
	
	var fullCasterSlots json.RawMessage 
	var warlockSlots json.RawMessage 
	
	if warlockLevel > 0 { 
		warlockSlots, err = cfg.DB.GetSpellSlotsMax(r.Context(), database.GetSpellSlotsMaxParams{ 
			CasterType: "warlock", 
			CasterLevel: int32(warlockLevel), 
		}) 
		if err != nil { 
			respondWithError(w, 500, "Error getting warlock spell slots") 
			return 
		} 
	} 
	
	effectiveCasterLevel, err := cfg.DB.GetCasterLevel(r.Context(), input.ID) 
	if err != nil { 
		respondWithError(w, 500, "Error getting effective caster level") 
		return 
	} 
	
	if effectiveCasterLevel > 0 { 
		fullCasterSlots, err = cfg.DB.GetSpellSlotsMax(r.Context(), database.GetSpellSlotsMaxParams{ 
			CasterType: "full", 
			CasterLevel: int32(effectiveCasterLevel), 
		}) 
		if err != nil { 
			respondWithError(w, 500, "Error getting full caster spell slots") 
			return 
		} 
	} 

	resp := SpellSlotsResponse{ 
		FullCasterSlots: fullCasterSlots, 
		WarlockSlots: warlockSlots, 
	} 
	
	respondWithJSON(w, 200, resp)
	return
}

func (cfg *APIConfig) HandlerGetUserCharacters(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header) 
	if err != nil { 
		respondWithError(w, 401, "Error retrieving token") 
		return 
	} 
	
	userID, _, err := auth.ValidateJWT(token, cfg.Secret) 
	if err != nil { 
		respondWithError(w, 401, "Error validating token") 
		return 
	}

	characters, err := cfg.DB.GetUserCharacters(r.Context(), userID)
	if err != nil {
		respondWithError(w, 500, "Error getting characters")
		return
	}

	returnSlice := []Character{}

	for _, character := range characters {
		val := Character{
			ID:				character.ID,
			Name:			character.Name,
			ClassLevels:	character.ClassLevels,
		}
		returnSlice = append(returnSlice, val)
	}

	respondWithJSON(w, 200, returnSlice)
	return
}

func (cfg *APIConfig) HandlerCharacterSpells(w http.ResponseWriter, r *http.Request) {
	type Input struct {
		Index		string		`json:"index"`
		ID			uuid.UUID	`json:"id"`
		Name		string		`json:"name"`
	}

	token, err := auth.GetBearerToken(r.Header) 
	if err != nil { 
		respondWithError(w, 401, "Error retrieving token") 
		return 
	} 
	
	_, _, err = auth.ValidateJWT(token, cfg.Secret) 
	if err != nil { 
		respondWithError(w, 401, "Error validating token") 
		return 
	} 
	
	decoder := json.NewDecoder(r.Body)
    input := Input{}

    err = decoder.Decode(&input)
    if err != nil {
        respondWithError(w, 500, "Error decoding input")
        return
    }
	
	spellID, err := cfg.DB.GetSpellID(r.Context(), input.Index)
	if err != nil {
		respondWithError(w, 500, "Error getting spell ID")
		return
	}

	_, err = cfg.DB.AddCharacterSpell(r.Context(), database.AddCharacterSpellParams{
		SpellID:		spellID,
		CharID:			input.ID,
	})
	if err != nil {
		respondWithError(w, 500, "Error adding spell")
		return
	}

	respondWithMessage(w, 200, "Spell added")
	return
}

func (cfg *APIConfig) HandlerGetCharacterSpells(w http.ResponseWriter, r *http.Request) {
	type Input struct {
		ID			uuid.UUID	`json:"id"`
		Name		string		`json:"name"`
	}

	token, err := auth.GetBearerToken(r.Header) 
	if err != nil { 
		respondWithError(w, 401, "Error retrieving token") 
		return 
	} 
	
	_, _, err = auth.ValidateJWT(token, cfg.Secret) 
	if err != nil { 
		respondWithError(w, 401, "Error validating token") 
		return 
	} 
	
	decoder := json.NewDecoder(r.Body)
    input := Input{}

    err = decoder.Decode(&input)
    if err != nil {
        respondWithError(w, 500, "Error decoding input")
        return
    }

	charSpells, err := cfg.DB.GetCharacterSpells(r.Context(), input.ID)
	if err != nil {
		respondWithError(w, 500, "Error getting spells")
		return
	}

	returnSlice := []SpellNameUrl{}

	for _, spell := range charSpells {
		val := SpellNameUrl{
			Index:		spell.Index,
			Name:		spell.Name,
			Level:		spell.Level.Int32,
			Url:		spell.Url,
		}
		returnSlice = append(returnSlice, val)
	}

	respondWithJSON(w, 200, returnSlice)
	return
}

func (cfg *APIConfig) HandlerRemoveCharacterSpell(w http.ResponseWriter, r *http.Request) {
	type Input struct {
		ID 		uuid.UUID 	`json:"id"`
		Index 	string		`json:"index"`
	}

	token, err := auth.GetBearerToken(r.Header) 
	if err != nil { 
		respondWithError(w, 401, "Error retrieving token") 
		return 
	} 
	
	_, _, err = auth.ValidateJWT(token, cfg.Secret) 
	if err != nil { 
		respondWithError(w, 401, "Error validating token") 
		return 
	} 
	
	decoder := json.NewDecoder(r.Body)
    input := Input{}

    err = decoder.Decode(&input)
    if err != nil {
        respondWithError(w, 500, "Error decoding input")
        return
    }

	spellID, err := cfg.DB.GetSpellID(r.Context(), input.Index)
	if err != nil {
		respondWithError(w, 500, "Error getting spell ID")
		return
	}

	err = cfg.DB.RemoveCharacterSpell(r.Context(), database.RemoveCharacterSpellParams{
		SpellID:	spellID,
		CharID:		input.ID,
	})
	if err != nil {
		respondWithError(w, 500, "Error removing spell from character")
		return
	}

	respondWithMessage(w, 201, "Spell removed")
	return
}
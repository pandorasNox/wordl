package models

import "testing"

func TestTemplateDataSuggest_validate(t *testing.T) {
	type fields struct {
		Word     string
		Action   string
		Language string
		Message  string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr error
	}{
		{name: "Suggested word match", fields: fields{Word: "gamer", Action: "add", Language: "english", Message: "test"}, wantErr: nil},
		{name: "Suggested word match", fields: fields{Word: "GAMER", Action: "add", Language: "english", Message: "test"}, wantErr: nil},
		{name: "Suggested word match", fields: fields{Word: "preuß", Action: "add", Language: "english", Message: "test"}, wantErr: nil},
		{name: "Suggested word match", fields: fields{Word: "höste", Action: "add", Language: "english", Message: "test"}, wantErr: nil},
		{name: "Suggested word match", fields: fields{Word: "hÖste", Action: "add", Language: "english", Message: "test"}, wantErr: nil},
		{name: "Suggested word match", fields: fields{Word: "HÖSTE", Action: "add", Language: "english", Message: "test"}, wantErr: nil},
		{name: "Suggested word match", fields: fields{Word: "fülle", Action: "add", Language: "english", Message: "test"}, wantErr: nil},
		{name: "Suggested word match", fields: fields{Word: "FÜLLE", Action: "add", Language: "english", Message: "test"}, wantErr: nil},
		{name: "Suggested word match", fields: fields{Word: "größe", Action: "add", Language: "english", Message: "test"}, wantErr: nil},
		{name: "Suggested word match", fields: fields{Word: "GRÖßE", Action: "add", Language: "english", Message: "test"}, wantErr: nil},

		{name: "Suggested word invalid (special chars: ?)", fields: fields{Word: "?????"}, wantErr: ErrFailedWordValidation},
		{name: "Suggested word invalid (special chars: ô)", fields: fields{Word: "grôss"}, wantErr: ErrFailedWordValidation},
		{name: "Suggested word invalid (special chars: emoji's (😁))", fields: fields{Word: "😁,😁,😁"}, wantErr: ErrFailedWordValidation},
		{name: "Suggested word invalid (word to short en)", fields: fields{Word: "tiny"}, wantErr: ErrFailedWordValidation},
		{name: "Suggested word invalid (word to short de)", fields: fields{Word: "kurz"}, wantErr: ErrFailedWordValidation},
		{name: "Suggested word invalid (word to long en)", fields: fields{Word: "toolong"}, wantErr: ErrFailedWordValidation},
		{name: "Suggested word invalid (word to long de)", fields: fields{Word: "zulang"}, wantErr: ErrFailedWordValidation},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tds := TemplateDataSuggest{
				Word:     tt.fields.Word,
				Action:   tt.fields.Action,
				Language: tt.fields.Language,
				Message:  tt.fields.Message,
			}
			if err := tds.Validate(); err != tt.wantErr {
				t.Errorf("TemplateDataSuggest.validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

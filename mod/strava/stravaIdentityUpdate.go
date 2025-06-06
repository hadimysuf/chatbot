package strava

import (
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocroot/helper/atdb"
	"github.com/whatsauth/itmodel"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func StravaIdentityUpdateHandler(Profile itmodel.Profile, Pesan itmodel.IteungMessage, db *mongo.Database) string {
	reply := "Informasi Profile Stava kakak: "

	if msg := maintenance(Pesan.Phone_number); msg != "" {
		reply += msg
		return reply
	}

	col := "strava_identity"
	// cek apakah akun strava sudah terdaftar di database
	data, err := atdb.GetOneDoc[StravaIdentity](db, col, bson.M{"phone_number": Pesan.Phone_number})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "Kak, akun Strava kamu belum terdaftar. Silakan daftar dulu!"
		}
		return "\n\nError fetching data dari MongoDB: " + err.Error()
	}
	if data.LinkIndentity == "" {
		return "link Strava kamu belum tersimpan di database!"
	}

	c := colly.NewCollector(
		colly.AllowedDomains(domWeb),
	)

	stravaIdentity := StravaIdentity{}
	stravaIdentity.AthleteId = data.AthleteId

	c.OnHTML("main", func(e *colly.HTMLElement) {
		stravaIdentity.Name = e.ChildText("h2[data-testid='details-name']")
		stravaIdentity.Picture = extractStravaProfileImg(e, stravaIdentity.Name)
	})

	c.OnScraped(func(r *colly.Response) {
		if data.AthleteId == "" {
			reply += "\n\nAkun Strava kak " + Pesan.Alias_name + " belum terdaftar."
			return
		}

		if stravaIdentity.Picture == "" {
			reply += "\n\nMaaf kak, sistem tidak dapat mengambil foto profil Strava kamu. Pastikan akun Strava kamu dibuat public(everyone). doc: https://www.do.my.id/mentalhealt-strava"
			return
		}

		if data.AthleteId == stravaIdentity.AthleteId {
			stravaIdentity.UpdatedAt = time.Now()

			updateData := bson.M{
				"picture":    stravaIdentity.Picture,
				"name":       stravaIdentity.Name,
				"updated_at": stravaIdentity.UpdatedAt,
			}

			// update data ke database jika ada perubahan
			_, err := atdb.UpdateDoc(db, col, bson.M{"athlete_id": stravaIdentity.AthleteId}, bson.M{"$set": updateData})
			if err != nil {
				reply += "\n\nError updating data to MongoDB: " + err.Error()
				return
			}

			reply += "\n\nData kak " + Pesan.Alias_name + " sudah berhasil di update."
			reply += "\n\nStrava Profile Picture: " + stravaIdentity.Picture
			reply += "\n\nCek link di atas apakah sudah sama dengan Strava Profile Picture di profile akun do.my.id yaa"

			conf, err := getConfigByPhone(db, Profile.Phonenumber)
			if err != nil {
				reply += "\n\nWah kak " + Pesan.Alias_name + " " + err.Error()
				return
			}

			dataToUser := map[string]interface{}{
				"stravaprofilepicture": stravaIdentity.Picture,
				"athleteid":            stravaIdentity.AthleteId,
				"phonenumber":          Pesan.Phone_number,
				"name":                 Pesan.Alias_name,
			}

			err = postToDomyikado(conf.DomyikadoSecret, conf.DomyikadoUserURL, dataToUser)
			if err != nil {
				reply += "\n\n" + err.Error()
				return
			}

			reply += "\n\nUpdate Strava Profile Picture berhasil dilakukan di do.my.id, silahkan cek di profile akun do.my.id kakak."

		} else {
			reply += "\n\nData Strava kak " + Pesan.Alias_name + " tidak ditemukan."
			return
		}
	})

	err = c.Visit(data.LinkIndentity)
	if err != nil {
		return "Link Profile Strava yang anda kirimkan tidak valid. Silakan kirim ulang dengan link yang valid.(3)"
	}

	return reply
}

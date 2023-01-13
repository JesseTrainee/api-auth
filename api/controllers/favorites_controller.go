package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/JesseTrainee/api-auth/api/auth"
	"github.com/JesseTrainee/api-auth/api/models"
	"github.com/JesseTrainee/api-auth/api/responses"
	"github.com/JesseTrainee/api-auth/api/utils/formaterror"
	"github.com/gorilla/mux"
)

func (server *Server) CreateFavorite(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	Favorite := models.Favorite{}
	err = json.Unmarshal(body, &Favorite)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	Favorite.Prepare()
	err = Favorite.Validate()
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	uid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}
	if uid != Favorite.UserID {
		responses.ERROR(w, http.StatusUnauthorized, errors.New(http.StatusText(http.StatusUnauthorized)))
		return
	}
	FavoriteCreated, err := Favorite.SaveFavorite(server.DB)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	w.Header().Set("Location", fmt.Sprintf("%s%s/%d", r.Host, r.URL.Path, FavoriteCreated.ID))
	responses.JSON(w, http.StatusCreated, FavoriteCreated)
}

func (server *Server) GetFavorites(w http.ResponseWriter, r *http.Request) {

	favorite := models.Favorite{}

	Favorites, err := favorite.FindAllFavorites(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	responses.JSON(w, http.StatusOK, Favorites)
}

func (server *Server) GetFavorite(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	fid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	favorite := models.Favorite{}

	favoriteReceived, err := favorite.FindFavoriteByID(server.DB, fid)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	responses.JSON(w, http.StatusOK, favoriteReceived)
}

func (server *Server) UpdateFavorite(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	// Check if the favorite id is valid
	fid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	//CHeck if the auth token is valid and  get the user id from it
	uid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	// Check if the favorite exist
	favorite := models.Favorite{}
	err = server.DB.Debug().Model(models.Favorite{}).Where("id = ?", fid).Take(&favorite).Error
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("Favorite not found"))
		return
	}

	// If a user attempt to update a favorite not belonging to him
	if uid != favorite.UserID {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}
	// Read the data favoriteed
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	// Start processing the request data
	favoriteUpdate := models.Favorite{}
	err = json.Unmarshal(body, &favoriteUpdate)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	//Also check if the request user id is equal to the one gotten from token
	if uid != favoriteUpdate.UserID {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	favoriteUpdate.Prepare()
	err = favoriteUpdate.Validate()
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	favoriteUpdate.ID = favorite.ID //this is important to tell the model the favorite id to update, the other update field are set above

	favoriteUpdated, err := favoriteUpdate.UpdateAFavorite(server.DB)

	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	responses.JSON(w, http.StatusOK, favoriteUpdated)
}

func (server *Server) DeleteFavorite(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	// Is a valid favorite id given to us?
	fid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	// Is this user authenticated?
	uid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	// Check if the favorite exist
	favorite := models.Favorite{}
	err = server.DB.Debug().Model(models.Favorite{}).Where("id = ?", fid).Take(&favorite).Error
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("Unauthorized"))
		return
	}

	// Is the authenticated user, the owner of this favorite?
	if uid != favorite.UserID {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}
	_, err = favorite.DeleteAFavorite(server.DB, fid, uid)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	w.Header().Set("Entity", fmt.Sprintf("%d", fid))
	responses.JSON(w, http.StatusNoContent, "")
}

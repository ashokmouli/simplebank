package gapi

import (
	"context"
	"database/sql"

	db "github.com/ashokmouli/simplebank/db/sqlc"
	"github.com/ashokmouli/simplebank/db/util"
	"github.com/ashokmouli/simplebank/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)
func (server *Server) LoginUser(ctx context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {

	// Fetch the user object from DB
	user, err := server.store.GetUser(ctx, req.GetUsername())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user does not exist: %s", err)
		}
		return nil, status.Errorf(codes.Internal, "internal error: %s", err)
	}

	// Check the password.
	err = util.CheckPassword(req.Password, user.HashedPassword)
	if err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "password mismatch: %s", err)
	}

	// Create access token.
	accessToken, accessPayload, err := server.maker.CreateToken(req.GetUsername(), server.config.AccessTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not create access token error: %s", err)
	}

	// Create refresh token.
	refresh_token, refreshPayload, err := server.maker.CreateToken(req.Username, server.config.RefreshTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not create refresh token error: %s", err)
	}

	metaData := extractMetaData(ctx)

	// Create a session record.
	session, err := server.store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshPayload.ID,
		Username:     req.GetUsername(),
		IsBlocked:    false,
		ClientIp:     sql.NullString{String: metaData.userAgent, Valid: true},
		UserAgent:    sql.NullString{String: metaData.clientIP, Valid: true},
		RefreshToken: refresh_token,
		ExpiresAt:    refreshPayload.ExpiredAt,
	})

	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not create session error: %s", err)
	}

	// Marshal back the response.

	rsp := &pb.LoginUserResponse{
		SessionID: session.ID.String(),
		AccessToken: accessToken,
		AccessTokenExpiresAt: timestamppb.New(accessPayload.ExpiredAt),
		RefreshToken: refresh_token,
		RefreshTokenExpiresAt: timestamppb.New(refreshPayload.ExpiredAt),
		User: convertUser(user),
	}
	return rsp, nil
}
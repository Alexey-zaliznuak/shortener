package handler

import (
	"context"
	"fmt"

	"github.com/Alexey-zaliznuak/shortener/internal/handler/protobuf"
	"github.com/Alexey-zaliznuak/shortener/internal/model"
	"github.com/Alexey-zaliznuak/shortener/internal/repository"
	"github.com/Alexey-zaliznuak/shortener/internal/repository/database"
	"github.com/Alexey-zaliznuak/shortener/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ShortenerService struct {
	protobuf.UnimplementedShortenerServiceServer

	service *service.LinksService
	auth    service.AuthService
}

func NewShortenerGRPCService(linksService *service.LinksService, authService *service.AuthService) *ShortenerService {
	return &ShortenerService{service: linksService, auth: *authService}
}

func (server *ShortenerService) ShortenURL(context context.Context, in *protobuf.URLShortenRequest) (*protobuf.URLShortenResponse, error) {
	var response protobuf.URLShortenResponse

	dto := &model.CreateLinkDto{FullURL: in.GetUrl()}
	claims, err := server.extractAuth(context)

	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	link, _, err := server.service.CreateLink(dto, claims.UserID)

	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	response.SetResult(link.Shortcut)

	return &response, nil
}

func (server *ShortenerService) ExpandURL(ctx context.Context, in *protobuf.URLExpandRequest) (*protobuf.URLExpandResponse, error) {
	var response protobuf.URLExpandResponse

	shortcut := in.GetId()
	fullURL, err := server.service.GetFullURLFromShort(shortcut)

	if err != nil {
		if err == database.ErrObjectDeleted {
			return nil, status.Error(codes.NotFound, "ссылка удалена")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	response.SetResult(fullURL)

	return &response, nil
}

func (server *ShortenerService) ListUserURLs(ctx context.Context, in *emptypb.Empty) (*protobuf.UserURLsResponse, error) {
	var response protobuf.UserURLsResponse

	claims, err := server.extractAuth(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	links, err := server.service.GetUserLinksByUserID(claims.UserID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	urlDataList := make([]*protobuf.URLData, 0, len(links))
	for _, link := range links {
		urlData := &protobuf.URLData{}
		urlData.SetShortUrl(link.Shortcut)
		urlData.SetOriginalUrl(link.FullURL)
		urlDataList = append(urlDataList, urlData)
	}

	response.SetUrl(urlDataList)

	return &response, nil
}

func (server *ShortenerService) extractAuth(ctx context.Context) (*repository.Claims, error) {
	var token string

	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		values := md.Get("authorization")
		if len(values) > 0 {
			token = values[0]
		}
		return server.auth.Repository.ParsePayload(token)
	} else {
		return nil, fmt.Errorf("ошибка парсинга jwt токена")
	}
}

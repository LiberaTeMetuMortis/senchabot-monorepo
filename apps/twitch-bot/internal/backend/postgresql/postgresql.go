package postgresql

import (
	"context"
	"errors"

	"github.com/senchabot-dev/monorepo/apps/twitch-bot/internal/backend"
	"github.com/senchabot-dev/monorepo/apps/twitch-bot/internal/models"
	"gorm.io/gorm"
)

type PostgreSQLBackend struct {
	DB *gorm.DB
}

func NewPostgreSQLBackend(db *gorm.DB) backend.Backend {
	return &PostgreSQLBackend{
		DB: db,
	}
}

func (b *PostgreSQLBackend) GetTwitchChannels(ctx context.Context) ([]*models.TwitchChannel, error) {
	var twitchChannels []*models.TwitchChannel

	result := b.DB.Find(&twitchChannels)
	if result.Error != nil {
		return nil, errors.New("(GetTwitchChannels) db.Find Error:" + result.Error.Error())
	}

	return twitchChannels, nil
}

func (b *PostgreSQLBackend) CreateTwitchChannel(ctx context.Context, channelId string, channelName string, userId *string) (bool, error) {
	var twitchChannel []models.TwitchChannel

	result := b.DB.Where("channel_id = ?", channelId).Where("channel_name = ?", channelName).Find(&twitchChannel)
	if result.Error != nil {
		return false, errors.New("(CreateTwitchChannel) db.Find Error:" + result.Error.Error())
	}
	if len(twitchChannel) > 0 {
		return true, nil
	}

	twitchChannel = append(twitchChannel, models.TwitchChannel{
		ChannelId:   channelId,
		ChannelName: channelName,
		UserId:      nil,
	})

	result = b.DB.Create(&twitchChannel)
	if result.Error != nil {
		return false, errors.New("(CreateTwitchChannel) db.Exec Error:" + result.Error.Error())
	}

	return false, nil
}

func (b *PostgreSQLBackend) GetTwitchBotConfig(ctx context.Context, twitchChannelId string, configKey string) (*models.TwitchBotConfig, error) {
	var twitchBotConfig models.TwitchBotConfig
	result := b.DB.Where("twitch_channel_id = ?", twitchChannelId).Where("key = ?", configKey).First(&twitchBotConfig)

	if result.Error != nil {
		return nil, errors.New("(GetTwitchBotConfig) db.First Error:" + result.Error.Error())
	}

	return &twitchBotConfig, nil
}

func (b *PostgreSQLBackend) GetBotCommand(ctx context.Context, commandName string, twitchChannelId string) (*models.BotCommand, error) {
	var botCommand models.BotCommand

	result := b.DB.Where("command_name = ?", commandName).Where("twitch_channel_id = ?", twitchChannelId).First(&botCommand)
	if result.Error != nil {
		return nil, errors.New("(GetBotCommand) db.First Error:" + result.Error.Error())
	}

	return &botCommand, nil
}

func (b *PostgreSQLBackend) CreateBotCommand(ctx context.Context, commandName string, commandContent string, twitchChannelId string, createdBy string) (*string, error) {
	var botCommand []models.BotCommand
	var commandExistString string = "command_exists"

	commandExist, err := b.CheckCommandExists(ctx, commandName, twitchChannelId)
	if err != nil {
		return nil, err
	}
	if commandExist {
		return &commandExistString, nil
	}

	existAlias, err := b.CheckCommandAlias(ctx, commandName, twitchChannelId)
	if err != nil {
		return nil, err
	}

	if existAlias != nil {
		return existAlias, nil
	}

	botCommand = append(botCommand, models.BotCommand{
		CommandName:     commandName,
		CommandContent:  commandContent,
		TwitchChannelID: twitchChannelId,
		CreatedBy:       &createdBy,
	})

	result := b.DB.Create(&botCommand)
	if result.Error != nil {
		return nil, errors.New("(CreateBotCommand) db.Create Error:" + result.Error.Error())
	}

	return nil, nil
}

func (b *PostgreSQLBackend) CheckCommandExists(ctx context.Context, commandName string, twitchChannelId string) (bool, error) {
	var botCommand []models.BotCommand

	result := b.DB.Where("command_name = ?", commandName).Where("twitch_channel_id", twitchChannelId).Find(&botCommand)
	if result.Error != nil {
		return false, errors.New("(CheckCommandExists) db.Find Error:" + result.Error.Error())
	}
	if len(botCommand) > 0 {
		return true, nil
	}

	return false, nil
}

func (b *PostgreSQLBackend) UpdateBotCommand(ctx context.Context, commandName string, commandContent string, twitchChannelId string, updatedBy string) error {
	var botCommand *models.BotCommand

	result := b.DB.Where("command_name = ?", commandName).Where("twitch_channel_id = ?", twitchChannelId).First(&botCommand)
	if result.Error != nil {
		return errors.New("(UpdateBotCommand) db.Find Error:" + result.Error.Error())
	}

	result = b.DB.Model(&botCommand).Updates(models.BotCommand{
		CommandContent: commandContent,
		UpdatedBy:      &updatedBy,
	})
	if result.Error != nil {
		return errors.New("(UpdateBotCommand) db.Update Error:" + result.Error.Error())
	}

	return nil
}

func (b *PostgreSQLBackend) DeleteBotCommand(ctx context.Context, commandName string, twitchChannelId string) error {
	var botCommand *models.BotCommand

	result := b.DB.Where("command_name = ?", commandName).Where("twitch_channel_id = ?", twitchChannelId).First(&botCommand)
	if result.Error != nil {
		return errors.New("(DeleteBotCommand) db.First Error:" + result.Error.Error())
	}

	result = b.DB.Delete(&botCommand)
	if result.Error != nil {
		return errors.New("(DeleteBotCommand) db.Delete Error:" + result.Error.Error())
	}

	return nil
}

func (b *PostgreSQLBackend) CreateBotActionActivity(ctx context.Context, botPlatformType string, botActivity string, twitchChannelId string, commandAuthor string) error {
	botActionActivity := models.BotActionActivity{
		BotPlatformType: botPlatformType,
		BotActivity:     botActivity,
		TwitchChannelID: &twitchChannelId,
		CommandAuthor:   &commandAuthor,
	}

	result := b.DB.Create(&botActionActivity)

	if result.Error != nil {
		return errors.New("(CreateBotActionActivity) db.Create Error:" + result.Error.Error())
	}

	return nil
}

func (b *PostgreSQLBackend) CreateCommandAliases(ctx context.Context, commandName string, aliases []string, twitchChannelId string, createdBy string) (*string, error) {
	commandAliases := []models.BotCommandAlias{}

	for _, commandAlias := range aliases {
		existAlias, err := b.CheckCommandAlias(ctx, commandAlias, twitchChannelId)
		if err != nil {
			return nil, err
		}

		if existAlias != nil {
			return existAlias, nil
		}

		commandAlias := models.BotCommandAlias{
			CommandAlias:    commandAlias,
			CommandName:     commandName,
			TwitchChannelID: &twitchChannelId,
			CreatedBy:       createdBy,
		}
		commandAliases = append(commandAliases, commandAlias)
	}

	err := b.DB.Save(&commandAliases).Error
	if err != nil {
		return nil, errors.New("(CreateCommandAliases) db.Save Error:" + err.Error())
	}

	return nil, nil
}

func (b *PostgreSQLBackend) GetCommandAlias(ctx context.Context, command string, twitchChannelId string) (*string, error) {
	var commandAlias models.BotCommandAlias

	err := b.DB.Where("command_alias = ?", command).Where("twitch_channel_id = ?", twitchChannelId).First(&commandAlias).Error
	if err != nil {
		return nil, errors.New("(GetCommandAlias) db.Find Error:" + err.Error())
	}

	return &commandAlias.CommandName, nil
}

func (b *PostgreSQLBackend) CheckCommandAlias(ctx context.Context, commandAlias string, twitchChannelId string) (*string, error) {
	var commandAliasModel []models.BotCommandAlias

	result := b.DB.Where("command_alias = ?", commandAlias).Where("twitch_channel_id", twitchChannelId).Find(&commandAliasModel)
	if result.Error != nil {
		return nil, errors.New("(CheckCommandAlias) db.Find Error:" + result.Error.Error())
	}

	if len(commandAliasModel) > 0 {
		return &commandAliasModel[0].CommandAlias, nil
	}

	return nil, nil
}

func (b *PostgreSQLBackend) DeleteCommandAlias(ctx context.Context, commandAlias string, twitchChannelId string) error {
	var commandAliasModel *models.BotCommandAlias

	result := b.DB.Where("command_alias = ?", commandAlias).Where("twitch_channel_id = ?", twitchChannelId).First(&commandAliasModel)
	if result.Error != nil {
		return errors.New("(DeleteCommandAlias) db.First Error:" + result.Error.Error())
	}

	result = b.DB.Delete(&commandAliasModel)
	if result.Error != nil {
		return errors.New("(DeleteCommandAlias) db.Delete Error:" + result.Error.Error())
	}

	return nil
}

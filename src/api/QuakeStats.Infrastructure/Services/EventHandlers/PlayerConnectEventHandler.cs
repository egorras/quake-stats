using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Logging;
using QuakeStats.Domain.Events;
using QuakeStats.Infrastructure.Data;

namespace QuakeStats.Infrastructure.Services.EventHandlers;

public class PlayerConnectEventHandler(
    ApplicationDbContext dbContext,
    ILogger<PlayerConnectEventHandler> logger) : BaseEventHandler<PlayerConnectEvent>
{
    public override string EventType => "PLAYER_CONNECT";

    protected override async Task ProcessEventAsync(PlayerConnectEvent @event)
    {
        logger.LogInformation("Processing player connect event for player {SteamId}", @event.SteamId);

        var player = await dbContext.Players
            .FirstOrDefaultAsync(p => p.SteamId == @event.SteamId);

        if (player == null)
        {
            player = new();
            dbContext.Players.Add(player);
        }

        player.Apply(@event);
    }
}

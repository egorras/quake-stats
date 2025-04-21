using QuakeStats.Domain.Events;

namespace QuakeStats.Domain.Entities;

public class Player : BaseEntity
{
    public string Name { get; set; } = string.Empty;
    public string SteamId { get; set; } = string.Empty;

    public void Apply(PlayerConnectEvent @event)
    {
        Name = @event.Name;
        SteamId = @event.SteamId;
    }

    public void Apply(PlayerDisconnectEvent @event)
    {
        Name = @event.Name;
        SteamId = @event.SteamId;
    }
}

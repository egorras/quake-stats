using System.Text.Json;
using QuakeStats.Domain.Entities;
using QuakeStats.Domain.Events;
using QuakeStats.Domain.Services;

namespace QuakeStats.Infrastructure.Services.EventHandlers;

/// <summary>
/// Base implementation for event handlers
/// </summary>
/// <typeparam name="TEvent">The type of event this handler processes</typeparam>
public abstract class BaseEventHandler<TEvent> : IEventHandler where TEvent : BaseEvent
{
    /// <summary>
    /// Gets the event type string this handler can process
    /// </summary>
    public abstract string EventType { get; }

    /// <summary>
    /// Checks if this handler can process the given event type
    /// </summary>
    public virtual bool CanHandle(string eventType) => eventType == EventType;

    /// <summary>
    /// Processes an event
    /// </summary>
    public async Task HandleAsync(Event @event)
    {
        // Parse the event data into the specific event type
        var typedEvent = ParseEventData(@event.EventData);

        // Process the typed event
        await ProcessEventAsync(typedEvent);
    }

    private static readonly JsonSerializerOptions JsonSerializerOptions = new()
    {
        PropertyNameCaseInsensitive = true,
        PropertyNamingPolicy = JsonNamingPolicy.SnakeCaseUpper
    };

    /// <summary>
    /// Parse the event data JSON into the specific event type
    /// </summary>
    protected virtual TEvent ParseEventData(string eventData)
    {
        return JsonSerializer.Deserialize<TEvent>(eventData, JsonSerializerOptions)
            ?? throw new JsonException($"Failed to deserialize event data to {typeof(TEvent).Name}");
    }

    /// <summary>
    /// Process the typed event
    /// </summary>
    protected abstract Task ProcessEventAsync(TEvent @event);
}

using QuakeStats.Domain.Entities;
using QuakeStats.Domain.Events;

namespace QuakeStats.Domain.Services;

/// <summary>
/// Base interface for all event handlers
/// </summary>
public interface IEventHandler
{
    /// <summary>
    /// Gets the event type string this handler can process
    /// </summary>
    string EventType { get; }

    /// <summary>
    /// Checks if this handler can process the given event type
    /// </summary>
    bool CanHandle(string eventType);

    /// <summary>
    /// Processes an event
    /// </summary>
    Task HandleAsync(Event @event);
}

/// <summary>
/// Interface for event handlers that can process specific types of events
/// </summary>
/// <typeparam name="TEvent">The type of event this handler can process</typeparam>
public interface IEventHandler<TEvent> : IEventHandler where TEvent : BaseEvent
{
    /// <summary>
    /// Processes an event of the specific type
    /// </summary>
    Task HandleAsync(TEvent @event);
}
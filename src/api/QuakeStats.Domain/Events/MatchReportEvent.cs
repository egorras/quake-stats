using System.Text.Json.Serialization;
using QuakeStats.Domain.Enums;
using QuakeStats.Domain.Events.JsonConverters;

namespace QuakeStats.Domain.Events;

public record MatchReportEvent : BaseEvent
{
    public bool Aborted { get; set; }
    public int CaptureLimit { get; init; }
    public string? ExitMsg { get; init; }
    public string Factory { get; init; } = string.Empty;
    public string FactoryTitle { get; init; } = string.Empty;

    [JsonConverter(typeof(StringNullConverter))]
    public string? FirstScorer { get; init; }
    public int FragLimit { get; init; }
    public int GameLength { get; init; }
    public GameType GameType { get; init; }

    [JsonConverter(typeof(IntToBoolConverter))]
    public bool Infected { get; init; }

    [JsonConverter(typeof(IntToBoolConverter))]
    public bool Instagib { get; init; }
    public int LastLeadChangeTime { get; init; }

    [JsonConverter(typeof(StringNullConverter))]
    public string? LastScorer { get; init; }

    [JsonConverter(typeof(StringNullConverter))]
    public string? LastTeamScorer { get; init; }
    public string Map { get; init; } = null!;
    public int MercyLimit { get; init; }

    [JsonConverter(typeof(IntToBoolConverter))]
    public bool QuadHog { get; init; }

    [JsonConverter(typeof(IntToBoolConverter))]
    public bool Restarted { get; init; }
    public int RoundLimit { get; init; }
    public int ScoreLimit { get; init; }
    public string ServerTitle { get; init; } = string.Empty;
    public int TimeLimit { get; init; }

    [JsonConverter(typeof(IntToBoolConverter))]
    public bool Training { get; init; }
    public int TeamScoreRed { get; init; }
    public int TeamScoreBlue { get; init; }
}

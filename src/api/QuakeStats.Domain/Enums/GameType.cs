using System.ComponentModel.DataAnnotations;
using System.Text.Json.Serialization;

namespace QuakeStats.Domain.Enums;

[JsonConverter(typeof(JsonStringEnumConverter))]
public enum GameType
{
    [Display(Name = "Attack & Defense")]
    [JsonStringEnumMemberName("AD")]
    AttackAndDefense,

    [Display(Name = "Clan Arena")]
    [JsonStringEnumMemberName("CA")]
    ClanArena,

    [Display(Name = "Capture The Flag")]
    [JsonStringEnumMemberName("CTF")]
    CaptureTheFlag,

    [Display(Name = "Duel")]
    [JsonStringEnumMemberName("DUEL")]
    Duel,

    [Display(Name = "Domination")]
    [JsonStringEnumMemberName("DOM")]
    Domination,

    [Display(Name = "Free For All")]
    [JsonStringEnumMemberName("FFA")]
    FreeForAll,

    [Display(Name = "Freeze Tag")]
    [JsonStringEnumMemberName("FT")]
    FreezeTag,

    [Display(Name = "Harvester")]
    [JsonStringEnumMemberName("HAR")]
    Harvester,

    [Display(Name = "One-Flag CTF")]
    [JsonStringEnumMemberName("ONEFLAG")]
    OneFlag,

    [Display(Name = "Race")]
    [JsonStringEnumMemberName("RACE")]
    Race,

    [Display(Name = "Red Rover")]
    [JsonStringEnumMemberName("RR")]
    RedRover,

    [Display(Name = "Team Deathmatch")]
    [JsonStringEnumMemberName("TDM")]
    TeamDeathmatch
}


from typing import List, Dict, Optional
from enum import Enum
from copy import deepcopy, copy
import time
from itertools import groupby
from operator import itemgetter
import random
from functools import lru_cache
from colorama import init as colorama_init
colorama_init()
from colorama import Fore, Back, Style
from cfr import Game, Action as CFRAction, State as CFRState, CFR

HIDDEN_CARD = """\
┌─────────┐
│░░░░░░░░░│
│░░░░░░░░░│
│░░░░░░░░░│
│░░░░░░░░░│
│░░░░░░░░░│
│░░░░░░░░░│
│░░░░░░░░░│
└─────────┘
"""
BLANK_CARD = """\
┌─────────┐
│         │
│         │
│         │
│         │
│         │
│         │
│         │
└─────────┘
"""
PRINT_CARD = """\
┌─────────┐
│{}       │
│         │
│         │
│    {}   │
│         │
│         │
│       {}│
└─────────┘
""".format('{rank: <2}', '{suit: <2}', '{rank: >2}')


class Suit(Enum):
    Spades = 'Spades'
    Clubs = 'Clubs'
    Diamonds = 'Diamonds'
    Hearts = 'Hearts'

    def to_symbol(self) -> str:
        symbols = {
            'S':   '♠',
            'D': '♦',
            'H':   '♥',
            'C':    '♣',
        }
        return symbols[self.value]
        
class Value(Enum):
    Ace = 6
    King = 5
    Queen = 4
    Jack = 3
    Ten = 2
    Nine = 1

class Card:
    def __init__(self, value: Value, suit: Suit):
        self.value = value
        self.suit = suit

    def effective_suit(self, trump_suit):
        if self.value == Value.Jack and self.suit == SUIT_COMPLEMENT[trump_suit]:
            return trump_suit
        else:
            return self.suit

    def __hash__(self):
        return hash(self.get_code())

    def __eq__(self, other) -> bool:
        return self.value == other.value and self.suit == other.suit
    
    def __repr__(self) -> str:
        return f'{self.value.name} of {self.suit.name}'

    def __lt__(self, other) -> bool:
        if self.suit == other.suit:
            return self.value.value < other.value.value
        else:
            return self.suit.value < other.suit.value

    def get_code(self) -> str:
        return self.suit.value + str(self.value.value)

    def pretty_card(self, color=True) -> str:
        if self.value.value < 3:
            rank = str(self.value.value + 8)
        else:
            rank = self.value.name[0]

        output = PRINT_CARD.format(rank=rank, suit=self.suit.to_symbol())

        if color:
            
            if self.suit == Suit.Hearts:
                text_color = Fore.RED
            elif self.suit == Suit.Diamonds:
                text_color = Fore.YELLOW
            elif self.suit == Suit.Spades:
                text_color = Fore.LIGHTBLUE_EX
            else:
                text_color = Fore.GREEN
                
            output = '\n'.join([text_color + line + Fore.RESET for line in output.split('\n')])
                
        return output

class InvalidStateException(Exception):
    pass

class InvalidActionExection(Exception):
    pass

class Action(Enum):
    # Bidding actions
    CallDiamonds = 'CALL_Diamonds'
    CallHearts = 'CALL_Hearts'
    CallSpades = 'CALL_Spades'
    CallClubs = 'CALL_Clubs'
    PassBid = 'PASS_BID'

    # Play a card action
    PlayAD = 'PLAY_Ace_Diamonds'
    PlayKD = 'PLAY_King_Diamonds'
    PlayQD = 'PLAY_Queen_Diamonds'
    PlayJD = 'PLAY_Jack_Diamonds'
    PlayTD = 'PLAY_Ten_Diamonds'
    PlayND = 'PLAY_Nine_Diamonds'

    PlayAH = 'PLAY_Ace_Hearts'
    PlayKH = 'PLAY_King_Hearts'
    PlayQH = 'PLAY_Queen_Hearts'
    PlayJH = 'PLAY_Jack_Hearts'
    PlayTH = 'PLAY_Ten_Hearts'
    PlayNH = 'PLAY_Nine_Hearts'
    
    PlayAS = 'PLAY_Ace_Spades'
    PlayKS = 'PLAY_King_Spades'
    PlayQS = 'PLAY_Queen_Spades'
    PlayJS = 'PLAY_Jack_Spades'
    PlayTS = 'PLAY_Ten_Spades'
    PlayNS = 'PLAY_Nine_Spades'
    
    PlayAC = 'PLAY_Ace_Clubs'
    PlayKC = 'PLAY_King_Clubs'
    PlayQC = 'PLAY_Queen_Clubs'
    PlayJC = 'PLAY_Jack_Clubs'
    PlayTC = 'PLAY_Ten_Clubs'
    PlayNC = 'PLAY_Nine_Clubs'

    Discard0 = 0
    Discard1 = 1
    Discard2 = 2
    Discard3 = 3
    Discard4 = 4
    Discard5 = 5

    @staticmethod
    def get_bid_action(card: Card):
        return Action(f"CALL_{card.suit.name}")

def print_cards(cards: List[Card]):
    lines = zip(*[card.pretty_card().split('\n') for card in cards])
    print('\n'.join([' '.join(line) for line in lines]))

@lru_cache(16)
def get_card_rankings(trump_suit: Suit, lead_suit: Optional[Suit]) -> Dict[Card, int]:
    """
    Generates the card power rankings based on the current trump and lead suit.
    """
    if trump_suit is None:
        raise InvalidStateException('Cannot get card rankings when trump has not been called')
    if lead_suit is None:
        # Return a ranking so we can still group equivalent plays at the start of a trick
        ranking = []
        for suit in Suit:
                ranking += [Card(val, suit) for val in reversed(Value)]
        return {card: i for i, card in enumerate(ranking)}


    right_bower = Card(Value.Jack, trump_suit)
    left_bower = Card(Value.Jack, SUIT_COMPLEMENT[trump_suit])

    # Build the rankings in reverse order so the index serves as the value
    ranking = []

    # Add off-suits
    for suit in Suit:
        if suit != trump_suit and suit !=  lead_suit:
            ranking += [Card(val, suit) for val in reversed(Value)]
    
    if lead_suit != trump_suit:
        # Add lead suit
        ranking += [Card(val, lead_suit) for val in reversed(Value)]

    ranking += [Card(val, trump_suit) for val in reversed(Value)]

    # remove bowers
    ranking.remove(left_bower)
    ranking.remove(right_bower)

    ranking += [left_bower, right_bower]

    return {card: i for i, card in enumerate(ranking)}

ACTION_MAP = {
        Suit.Spades: {
            Value.Ace: Action.PlayAS,
            Value.King: Action.PlayKS,
            Value.Queen: Action.PlayQS,
            Value.Jack: Action.PlayJS,
            Value.Ten: Action.PlayTS,
            Value.Nine: Action.PlayNS
        },
        Suit.Diamonds: {
            Value.Ace: Action.PlayAD,
            Value.King: Action.PlayKD,
            Value.Queen: Action.PlayQD,
            Value.Jack: Action.PlayJD,
            Value.Ten: Action.PlayTD,
            Value.Nine: Action.PlayND
        },
        Suit.Hearts: {
            Value.Ace: Action.PlayAH,
            Value.King: Action.PlayKH,
            Value.Queen: Action.PlayQH,
            Value.Jack: Action.PlayJH,
            Value.Ten: Action.PlayTH,
            Value.Nine: Action.PlayNH
        },
        Suit.Clubs: {
            Value.Ace: Action.PlayAC,
            Value.King: Action.PlayKC,
            Value.Queen: Action.PlayQC,
            Value.Jack: Action.PlayJC,
            Value.Ten: Action.PlayTC,
            Value.Nine: Action.PlayNC
        }
    }

SUIT_COMPLEMENT = {
    Suit.Spades: Suit.Clubs,
    Suit.Clubs: Suit.Spades,
    Suit.Diamonds: Suit.Hearts,
    Suit.Hearts: Suit.Diamonds
}

class EuchreGame(Game):
    def get_num_players(self):
        return 4

class EuchreState(CFRState):
    """
    Team Even: Players 0 and 2
    Team Odd: Players 1 and 3
    """
    def __init__(self, debug=False):

        self.debug = debug
        # Deck
        self.played_cards: List[Card] = []
        self.player_hands: List[List[Card]]
        self.table: List[Card] = []
        self.kitty: List[Card] = []
        # Score
        self.team_tricks: List[int] = [0, 0]
        self.team_points: List[int] = [0, 0]
        
        # Hand state
        self.dealer: int = 3
        self.trump_suit: Optional[Suit] = None
        self.calling_team: Optional[int] = None
        
        # Immediate State
        self.lead_suit: Optional[Suit] = None
        self.lead: int = 0
        self.hand_turn: int = 0
        self.current_agent: int = 0

        # State for random sampling
        # Need to track if a player shows they are short suited
        #self.player_suits: List[List[Suit]]

        self._deal()

        # Deal with spades always as trump
        self.trump_suit = Suit.Spades
        self.dealer = random.randint(0,3)
        self.lead = (self.dealer + 1) % 4
        self.current_agent = self.lead

        # Skip tricks to test iteration
        for i in range(1):
            for play in range(4):
                self.take_action(random.choice(self.valid_actions()), copy=False)

    def __deepcopy__(self, memo):
        # Create new instance
        inst_copy = copy(self)

        # Deepcopy containers
        inst_copy.played_cards = self.played_cards[:]
        inst_copy.player_hands = [hand[:] for hand in self.player_hands]
        inst_copy.table = self.table[:]
        inst_copy.kitty = self.kitty[:]
        inst_copy.team_tricks = self.team_tricks[:]
        inst_copy.team_points = self.team_points[:]
        
        return inst_copy

    def get_current_agent(self) -> int:
        return self.current_agent

    def state_string(self):
        return  f'{self.lead_suit.value if self.lead_suit else "0"}_' + ''.join(card.get_code() for card in self.player_hands[self.current_agent])

    def is_terminal(self) -> bool:
        """
        If the calling team is set or has taken all the tricks, the hand is over
        """
        if self.calling_team is None:
            return False
        non_calling_team = 1 - self.calling_team
        return self.team_tricks[self.calling_team] == 5 or self.team_tricks[non_calling_team] == 3
            
    def _deal(self):
        deck = [Card(value, suit) for value in Value for suit in Suit]
        random.shuffle(deck)

        self.player_hands: List[List[Card]] = [
            deck[:5],
            deck[5:10],
            deck[10:15],
            deck[15:20]
        ]
        for i, _ in enumerate(self.player_hands):
            self.player_hands[i].sort()

        self.kitty: List[Card] = deck[20:]
        #self.player_suits = [[suit for suit in Suit] for _ in range(4)]

    def pretty_print_state(self):
        print(f'Team Even score: {self.team_points[0]}')
        print(f'Team Odd score: {self.team_points[1]}')
        print(f'Calling team {self.calling_team}')
        print(f'Even tricks {self.team_tricks[0]}, Odd tricks {self.team_tricks[1]}')
        print(f'Dealer: Player {self.dealer}')
        print(f'Current Player {self.current_agent}')
        # Calling
        if self.trump_suit is None:
            if self.hand_turn < 4:
                print(f'Current kitty card:\n{self.kitty[-1].pretty_card()}')

        # Playing
        else:
            print(f'Trump: {self.trump_suit}')
            print('Table:')
            print_cards(self.table)

        print('Your hand')
        hand = self.player_hands[self.current_agent]
        print_cards(hand)
       
    def _group_equivalent_cards(self, hand: List[Card], ranking: Dict[Card, int], trump_suit: Suit):
        hand_ranking = [(card, ranking[card]) for card in hand]
        hand_ranking.sort(key=lambda x: x[1])
        
        card_groups = []
        for k, g in groupby(
            enumerate(hand_ranking), key=lambda x: (x[0] - x[1][1], x[1][0].effective_suit(trump_suit))
        ):
            group = list(pair[0] for pair in map(itemgetter(1), g))
            card_groups.append(group)

        return card_groups

    def valid_actions(self) -> List[Action]:
        hand = self.player_hands[self.current_agent]
        if max(self.team_points) >= 10:
            return []
        elif self.trump_suit is None: # Bidding phase
            call_actions = {Action.CallClubs, Action.CallDiamonds, Action.CallHearts, Action.CallSpades}
            kitty_suit_action = Action.get_bid_action(self.kitty[-1])
            if self.hand_turn < 4:
                return [Action.PassBid, kitty_suit_action]
            elif self.hand_turn < 7:
                return [Action.PassBid, *call_actions.difference({kitty_suit_action})]
            elif self.hand_turn == 7:
                return [*call_actions.difference({kitty_suit_action})]
            else:
                raise InvalidStateException('There was an error, trump suit cannot be None after hand_turn 8')
        elif len(hand) > 5:
            return [act for act in Action if act.name.startswith('Discard')]
        else: # Playing the hand
            playable_cards: List[Card] = []  
            if self.lead_suit is not None:
                follow_lead_cards = [card for card in hand if card.effective_suit(self.trump_suit) == self.lead_suit]
                if any(follow_lead_cards):
                    # Must follow suit if possible
                    playable_cards = follow_lead_cards
                else:
                    playable_cards = hand
            else:
                # All cards are legal
                playable_cards = hand

            # Action abstraction
            grouped_cards = self._group_equivalent_cards(playable_cards, get_card_rankings(self.trump_suit, self.lead_suit), self.trump_suit)
            
            return [ACTION_MAP[group[-1].suit][group[-1].value] for group in grouped_cards]
            
    def take_action(self, action: Action, copy=True):
        if action not in self.valid_actions():
            raise InvalidActionExection(f'{action} is not valid in the current game state')

        if self.debug:
            print(f'Player {self.current_agent} takes {action}')

        if copy:
            new_state = deepcopy(self)
        else:
            new_state = self
        new_state.hand_turn += 1

        if action == Action.PassBid:
            # Bidding continues clockwise
            new_state.current_agent = (self.current_agent + 1) % 4
        elif action.name.startswith('Call'):
            new_state.trump_suit = Suit[action.value[5:]]
            new_state.calling_team = self.current_agent % 2

            if self.hand_turn < 4:
                # Dealer picks up card
                kitty_card = self.kitty[-1]
                new_state.player_hands[self.dealer] += [self.kitty[-1]]
                new_state.kitty.remove(kitty_card)
                new_state.current_agent = self.dealer
            else:
                # left of dealer starts
                new_state.current_agent = (self.dealer + 1) % 4
        elif action.name.startswith('Discard'):
            idx = action.value
            discarded_card = self.player_hands[self.current_agent][idx]
            new_state.player_hands[self.current_agent].remove(discarded_card)
            new_state.played_cards.append(discarded_card)
            new_state.current_agent = (self.dealer + 1) % 4
        elif action.name.startswith('Play'):
            _, value, suit = action.value.split('_')

            played_card = Card(Value[value], Suit[suit])
            new_state.player_hands[self.current_agent].remove(played_card)
            new_state.table.append(played_card)

            if self.lead_suit is None:
                if played_card == Card(Value.Jack, SUIT_COMPLEMENT[self.trump_suit]):
                    new_state.lead_suit = self.trump_suit
                else:
                    new_state.lead_suit = played_card.suit
            # elif played_card.suit != self.lead_suit:
            #     if self.lead_suit in new_state.player_suits[self.current_agent]:
            #         # The player must be shortsuited
            #         new_state.player_suits[self.current_agent].remove(self.lead_suit)
            
            # Handle trick end logic
            if len(new_state.table) == 4:
                rankings: Dict[Card, int] = get_card_rankings(new_state.trump_suit, new_state.lead_suit)
                trick: List[Card] = new_state.table.copy()
                winning_card: Card = max(trick, key=lambda card: rankings[card])
                winning_player = ((trick.index(winning_card) + self.lead) % 4)

                if self.debug:
                    print(f'Player {winning_player} takes the trick!')

                new_state.team_tricks[winning_player % 2] += 1
                new_state.lead = winning_player
                new_state.current_agent = winning_player

                # Clear table
                new_state.played_cards += new_state.table
                new_state.table = []
                new_state.lead_suit = None
            else:
                new_state.current_agent = (self.current_agent + 1) % 4

            # End of hand logic
            # if all(len(hand) == 0 for hand in new_state.player_hands):
            #     # award points
            #     non_calling_team = (new_state.calling_team + 1) % 2
            #     if new_state.team_tricks[new_state.calling_team] < new_state.team_tricks[non_calling_team]:
            #         new_state.team_points[non_calling_team] += 2
            #     elif new_state.team_tricks[new_state.calling_team] == 5:
            #         new_state.team_points[new_state.calling_team] += 2
            #     else:
            #         new_state.team_points[new_state.calling_team] += 1

            #     # deal new hand
            #     new_state._deal()
            #     new_state.trump_suit = None
            #     new_state.hand_turn = 0
            #     new_state.team_tricks = [0, 0]
            #     new_state.dealer = (new_state.dealer + 1) % 4
            #     new_state.lead = (new_state.dealer + 1) % 4
            #     new_state.current_agent = (new_state.dealer + 1) % 4

        return new_state

    def get_utility(self, player_id: int) -> float:
        player_team = player_id % 2
        non_calling_team = 1 - self.calling_team
        
        points = [0, 0]
        if self.team_tricks[non_calling_team] > self.team_tricks[self.calling_team]:
            points[non_calling_team] = 2
        elif self.team_tricks[self.calling_team] == 5:
            points[self.calling_team] = 2
        else:
            points[self.calling_team] = 1

        return points[player_team]

    def get_random_unseen(self):
        """
        Generate a new state by randomizing unseen cards
        """
        # The player whose point of view is being used to
        # randomize the cards. Everything they know will remain
        # true
        pov_player = 0
        new_state = deepcopy(self)

        deck = []
        # Get all unseen cards
        for i, hand in enumerate(new_state.player_hands):
            if i != pov_player:
                deck += hand
        deck += new_state.kitty[:3]

        # Need to fill the most restricted hands first
        hand_idxs = [(i, suits) for i, suits in enumerate(new_state.player_suits) if i != pov_player]
        hand_idxs.sort(key=lambda elem: len(elem[1]))

        for hand_idx, _ in hand_idxs:
            random.shuffle(deck)
            new_hand = []
            for card in deck:
                if card.suit in new_state.player_suits[hand_idx]:
                    new_hand.append(card)
                if len(new_hand) == len(self.player_hands[hand_idx]):
                    break
            new_state.player_hands[hand_idx] = new_hand
            
            for card in new_hand:
                deck.remove(card)

        assert(len(deck) == 3)
        new_state.kitty[:3] = deck

        return new_state

def simulate_uniform(state: EuchreState):
    starting_team = state.current_agent % 2
    opponent = (starting_team + 1) % 2
    starting_points = state.team_points.copy()
    while True:
        if starting_points != state.team_points:
            break
        else:
            action = random.choice(state.valid_actions())
            state = state.take_action(action, copy=False)
    
    positive_points = state.team_points[starting_team] - starting_points[starting_team]
    negative_points = state.team_points[opponent] - starting_points[opponent]

    return positive_points - negative_points


def score_actions(state: EuchreState, num_samples=1000):
    actions = state.valid_actions()
    if len(actions) == 1:
        return actions, [None]
    action_rewards = [0] * len(actions)
    for i, action in enumerate(actions):
        for sample in range(num_samples // len(actions)):
            perturb_state = state.get_random_unseen()
            perturb_state.debug = False
            reward = simulate_uniform(perturb_state)
            action_rewards[i] += reward

    return actions, [reward / (num_samples // len(actions)) for reward in action_rewards]


if __name__ == "__main__":

    cfr = CFR(EuchreGame, EuchreState)
    times = []
    for i in range(10):
        start = time.time()
        cfr.train(1)
        end = time.time()

        print('Total time for 1 iteration:', end-start)
        times.append(end-start)
    print('Mean time', sum(times)/10)
    # suit_map = {
    #     Suit.Clubs: Suit.Diamonds,
    #     Suit.Diamonds: Suit.Clubs,
    #     Suit.Spades: Suit.Hearts,
    #     Suit.Hearts: Suit.Spades
    # }
        
    # deck = [Card(value, suit) for value in Value for suit in Suit]

    # player_hands: List[List[Card]] = [
    #         deck[:5],
    #         deck[5:10],
    #         deck[10:15],
    #         deck[15:20]
    #     ]

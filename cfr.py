from typing import List, Dict, Type
from functools import reduce

class Game:
    def get_num_players(self) -> int:
        raise NotImplementedError()

class Action:
    pass

class Strategy:
    def __init__(
        self,
        valid_actions
    ):
        pass
    def sample_action(self) -> Action:
        pass


class InfoSet:
    def __init__(
        self,
        valid_actions: List[Action]
    ):
        self.cumulative_regrets: Dict[Action, float] = {action: 0.0 for action in valid_actions}
        self.cumulative_strategy: Dict[Action, float] = {action: 0.0 for action in valid_actions}

        self.current_strategy: Dict[Action, float] = {action: 1/len(valid_actions) for action in valid_actions}

    def update_current_strategy(self):
        normalization = 0
        for action, regret in self.cumulative_regrets:
            self.current_strategy[action] = regret if regret > 0 else 0
            normalization += self.current_strategy[action]

        if normalization > 0:
            for action in self.cumulative_regrets.keys():
                self.current_strategy[action] = self.current_strategy[action] / normalization
        else:
            self.current_strategy = {action: 1 / len(self.current_strategy)}

        
    def get_strategy_prob(self, action: Action) -> float:
        return self.current_strategy[action]

    def get_average_strategy_prob(self, action: Action) -> Strategy:
        pass


class State:
    def __init__(
        self,
        num_agents: int
    ):
        pass

    def get_current_agent(self) -> int:
        pass

    def get_info_set(self) -> InfoSet:
        pass

    def valid_actions(self) -> List[Action]:
        raise NotImplementedError()

    def take_action(self, action: Action) -> State:
        raise NotImplementedError()

    def is_terminal(self) -> bool:
        raise NotImplementedError()

    def get_utility(self, player_id: int) -> float:
        raise NotImplementedError()

    def state_string(self) -> str:
        """
        Should return a unique, minimal string for the current state
        """
        raise NotImplementedError()


class CFR:
    """
    Implements a generic CFR algorithm using the above abstract classes.
    """
    def __init__(
            self,
            game_cls: Type[Game],
            state_cls: Type[State]
        ):
        self.infoSet_map: Dict[str, InfoSet] = {}

        self._state_cls = state_cls
        self.game = game_cls()
        
    def _cfr(self, agent_id, state: State, agent_path_probs: List[float]):
        current_agent = state.get_current_agent()
        
        # Return terminal state payoff
        if state.is_terminal():
            return state.get_utility(agent_id)

        # Get info set
        info_set_key = state.state_string()
        info_set = self.infoSet_map.get(info_set_key)
        if info_set is None:
            info_set = state.get_info_set()

        utility = 0
        action_utils = {}
        for action in state.valid_actions():
            action_prob = info_set.get_strategy_prob(action)
            
            new_path_probs = agent_path_probs.copy()
            new_path_probs[current_agent] = new_path_probs[current_agent] * action_prob

            action_utils[action] = self._cfr(agent_id, state.take_action(action), new_path_probs)
            utility += action_prob * action_utils[action]

        if current_agent == agent_id:
            non_player_prob = 1.0
            for i, prob in enumerate(agent_path_probs):
                if i != agent_id:
                    non_player_prob *= prob

            for action in state.valid_actions():
                info_set.cumulative_regrets[action] += non_player_prob * (action_utils[action] - utility)
                info_set.cumulative_strategy[action] += agent_path_probs[agent_id] * info_set.get_strategy_prob(action)

            info_set.update_current_strategy()

        return utility


    def train(self, iterations: int):
        util = 0
        num_players = self.game.get_num_players()
        for i in range(iterations):
            for player in range(num_players):
                # Get random game init
                random_state = self._state_cls()
                util += self._cfr(player, random_state, [1.0] * num_players)
        print(f'average utility {util/iterations}')

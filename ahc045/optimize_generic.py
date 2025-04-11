import optuna
import subprocess
import re
import os
import json # 設定ファイルを使う場合 (今回はdictで)
from typing import Dict, Any, List
import sys # エラー出力用

# ==============================================================================
# 設定セクション (ここを修正して異なる問題に適応)
# ==============================================================================
CONFIG: Dict[str, Any] = {
    # --- Optuna Study Settings ---
    "study_name": "ahc045_tuning", # Studyの名前
    "direction": "minimize",  # "maximize" または "minimize"
    "n_trials": 100,         # 試行回数

    # --- Command Execution Settings ---
    "command_template": 'fj test "{solver_command}" -n 50 -p 4', # 実行するコマンド全体。{solver_command} が置き換えられる
    "solver_executable": "./bin/a.out",  # 実行したいソルバー（Goプログラムなど）を起動するコマンド
    "param_prefix": "-",         # パラメータ名の前につける文字 (例: "-", "--")
    "param_separator": " ",      # パラメータ名と値の間の文字 (例: " ", "=")

    # --- Score Parsing Settings ---
    # fjコマンドが数値のみを出力する場合、この正規表現は objective 関数内では使用されないが、
    # 設定の汎用性のために残しておくことも可能。
    "score_regex": r"Score:\s*([+-]?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?)",
    "error_value": None,      # エラー時の代替値 (directionに応じて自動設定される)

    # --- Hyperparameter Definitions ---
    "hyperparameters": [
        {
            "name": "pram1",  # Goのflag名に合わせる
            "type": "float",           # データ型
            "low": 1.00,               # 指定された最小値
            "high": 1.20,              # 指定された最大値
            "step": 0.001,             # ステップ幅 (floatの場合)
        },
        {
            "name": "pram2",   # Goのflag名に合わせる
            "type": "float",           # データ型
            "low": 0.60,               # 指定された最小値
            "high": 1.00,              # 指定された最大値
            "step": 0.001,             # ステップ幅 (floatの場合)
        },
        {
            "name": "pram3",   # Goのflag名に合わせる
            "type": "float",           # データ型
            "low": 0.90,               # 指定された最小値
            "high": 1.00,               # 指定された最大値
            "step": 0.001,             # ステップ幅 (floatの場合)
        },
        {
            "name": "pram4",   # Goのflag名に合わせる
            "type": "float",           # データ型
            "low": 0.90,               # 指定された最小値
            "high": 1.00,               # 指定された最大値
            "step": 0.001,             # ステップ幅 (floatの場合)
        },
        {
            "name": "pram5",   # Goのflag名に合わせる
            "type": "float",           # データ型
            "low": 0.3,               # 指定された最小値
            "high": 1.00,               # 指定された最大値
            "step": 0.001,             # ステップ幅 (floatの場合)
        }
    ]
}
# ==============================================================================

def validate_config(config: Dict[str, Any]):
    """設定の基本的な妥当性をチェック"""
    print("Validating configuration...")
    if config["direction"] not in ["maximize", "minimize"]:
        raise ValueError("Config Error: 'direction' must be 'maximize' or 'minimize'")

    if not isinstance(config["n_trials"], int) or config["n_trials"] <= 0:
        raise ValueError("Config Error: 'n_trials' must be a positive integer")

    if "{solver_command}" not in config["command_template"]:
         print("Warning: 'command_template' does not contain '{solver_command}'. Parameters might not be passed correctly.", file=sys.stderr)

    if not config["solver_executable"]:
        raise ValueError("Config Error: 'solver_executable' cannot be empty")

    # score_regex は使わない場合もあるので、必須チェックはしないか、条件付きにする
    # try:
    #     re.compile(config["score_regex"])
    # except re.error as e:
    #     raise ValueError(f"Config Error: 'score_regex' is invalid: {e}")

    if not isinstance(config["hyperparameters"], list):
         raise ValueError("Config Error: 'hyperparameters' must be a list")

    if not config["hyperparameters"]:
         print("Warning: 'hyperparameters' list is empty. No parameters will be tuned.", file=sys.stderr)

    for i, param in enumerate(config["hyperparameters"]):
        if not isinstance(param, dict):
             raise ValueError(f"Config Error: Hyperparameter definition at index {i} must be a dictionary")
        if "name" not in param or not isinstance(param["name"], str) or not param["name"]:
            raise ValueError(f"Config Error: Hyperparameter at index {i} must have a non-empty 'name' (string)")
        if "type" not in param or param["type"] not in ["float", "int", "categorical", "loguniform"]:
            raise ValueError(f"Config Error: Hyperparameter '{param['name']}' has invalid 'type'. Must be one of: float, int, categorical, loguniform")

        ptype = param["type"]
        required_keys = []
        if ptype in ["float", "int", "loguniform"]:
            required_keys = ["low", "high"]
        elif ptype == "categorical":
            required_keys = ["choices"]

        for key in required_keys:
            if key not in param:
                raise ValueError(f"Config Error: Hyperparameter '{param['name']}' (type: {ptype}) is missing required key '{key}'")

        # 型チェック (より詳細に)
        if ptype == "categorical" and (not isinstance(param["choices"], list) or not param["choices"]):
             raise ValueError(f"Config Error: Hyperparameter '{param['name']}' (type: categorical) must have a non-empty list for 'choices'")
        if ptype in ["float", "int", "loguniform"]:
             low, high = param.get("low"), param.get("high")
             if not (isinstance(low, (int, float)) and isinstance(high, (int, float))):
                  raise ValueError(f"Config Error: Hyperparameter '{param['name']}' needs numeric 'low' and 'high' values")
             if low >= high:
                  raise ValueError(f"Config Error: Hyperparameter '{param['name']}' needs 'low' to be strictly less than 'high'")
        # Optional step/log validation can be added here if needed

    print("Configuration validation successful.")


def get_error_value(direction: str) -> float:
    """最適化方向に応じたエラー時の値を返す"""
    if direction == "maximize":
        return float('-inf')
    elif direction == "minimize":
        return float('inf')
    else:
        # validate_config でチェック済みのはずだが念のため
        raise ValueError(f"Invalid direction '{direction}' encountered in get_error_value")


def objective(trial: optuna.trial.Trial, config: Dict[str, Any]) -> float:
    """汎用的なOptuna目的関数"""
    suggested_params: Dict[str, Any] = {}
    param_args: List[str] = []

    # 1. 設定に基づきハイパパラメータを提案
    for param_def in config["hyperparameters"]:
        name = param_def["name"]
        ptype = param_def["type"]

        # Optunaのsuggest APIを動的に呼び出す
        if ptype == "float":
            # suggest_float specific arguments
            step = param_def.get("step")
            log = param_def.get("log", False) # suggest_loguniformは別type推奨だが互換性のため
            val = trial.suggest_float(name, param_def["low"], param_def["high"], step=step, log=log)
        elif ptype == "int":
            # suggest_int specific arguments
            step = param_def.get("step", 1) # Default step for int is 1
            log = param_def.get("log", False)
            val = trial.suggest_int(name, param_def["low"], param_def["high"], step=step, log=log)
        elif ptype == "categorical":
            # suggest_categorical specific arguments
            val = trial.suggest_categorical(name, param_def["choices"])
        elif ptype == "loguniform":
             # suggest_float with log=True
             val = trial.suggest_float(name, param_def["low"], param_def["high"], log=True)
        else:
            # validate_config でチェック済みだが念のため
            raise ValueError(f"Internal Error: Unsupported parameter type '{ptype}' encountered during suggestion.")

        suggested_params[name] = val
        # コマンドライン引数形式に整形してリストに追加
        param_args.append(f"{config['param_prefix']}{name}{config['param_separator']}{val}")

    # 2. 実行コマンドを構築
    #    パラメータ部分を結合
    solver_param_string = " ".join(param_args)
    #    ソルバー実行コマンド部分を作成 (実行ファイル + パラメータ)
    solver_command = f"{config['solver_executable']} {solver_param_string}".strip()
    #    全体のコマンドテンプレートに埋め込む
    full_command = config["command_template"].format(solver_command=solver_command)

    print(f"\n----- Trial {trial.number} -----")
    print(f"Suggested params: {suggested_params}")
    print(f"Executing command: {full_command}")

    try:
        # 3. コマンドを実行し、結果を取得
        result = subprocess.run(
            full_command,
            shell=True,        # 文字列全体をシェル経由で実行する場合にTrue
            check=True,        # エラー終了時(exit code != 0)にCalledProcessErrorを送出
            capture_output=True,# 標準出力・標準エラーをキャプチャする
            text=True,         # 出力/エラーをテキスト(文字列)としてデコードする
            encoding='utf-8'   # デコード時のエンコーディング指定
        )
        output = result.stdout
        # 標準エラーも確認したい場合
        # stderr_output = result.stderr
        # if stderr_output:
        #    print(f"Trial {trial.number}: Stderr:\n{stderr_output[:500]}...")

        print(f"Trial {trial.number}: Stdout preview: {output[:100].strip()}...") # 出力確認用

        # 4. スコアを抽出 (fjコマンドは数値のみを出力する想定)
        try:
            # 標準出力全体を文字列として取得し、前後の空白を除去
            score_str = output.strip()
            # 文字列が空でないかチェック
            if not score_str:
                print(f"Trial {trial.number}: Error - Command output was empty.", file=sys.stderr)
                return config["error_value"] # エラー時の値を返す
            # 文字列をfloatに直接変換
            score = float(score_str)
            print(f"Trial {trial.number}: Parsed Score: {score}")
            return score
        except ValueError:
            # floatへの変換に失敗した場合 (数値以外の出力があった場合など)
            print(f"Trial {trial.number}: Error - Output could not be converted to float.", file=sys.stderr)
            print(f"Output received: '{output}'", file=sys.stderr) # 実際の出力を確認用に表示
            return config["error_value"] # エラー時の値を返す

    except subprocess.CalledProcessError as e:
        # コマンド実行自体が失敗した場合 (例: ファイルが見つからない, Goコードがエラー終了)
        print(f"Trial {trial.number}: Error executing command: {e}", file=sys.stderr)
        print(f"Exit Code: {e.returncode}", file=sys.stderr)
        print(f"Stdout:\n{e.stdout[:500].strip()}...", file=sys.stderr) # 失敗時の標準出力
        print(f"Stderr:\n{e.stderr[:500].strip()}...", file=sys.stderr) # 失敗時の標準エラー
        return config["error_value"] # エラー時の値を返す

    except Exception as e:
        # その他の予期せぬエラー
        print(f"Trial {trial.number}: An unexpected error occurred: {e}", file=sys.stderr)
        import traceback
        traceback.print_exc(file=sys.stderr) # スタックトレースを出力
        return config["error_value"] # エラー時の値を返す

# --- Optuna Studyの作成と実行 ---
if __name__ == "__main__":
    print("Starting Optuna optimization script...")

    # 1. 設定の検証 (オプション)
    try:
        validate_config(CONFIG)
    except ValueError as e:
        print(f"\nConfiguration Error: {e}", file=sys.stderr)
        sys.exit(1) # エラーがあれば終了

    # 2. エラー時のデフォルト値を設定
    if CONFIG.get("error_value") is None:
        CONFIG["error_value"] = get_error_value(CONFIG["direction"])
        print(f"Setting default error value based on direction '{CONFIG['direction']}': {CONFIG['error_value']}")

    # 3. 実行ファイル/ディレクトリの存在チェック (簡易版、必要に応じて)
    #    solver_executable の内容に応じて適切なパスをチェックする
    #    例: exec_path_check = CONFIG["solver_executable"].split()[1] if CONFIG["solver_executable"].startswith("go run ") else CONFIG["solver_executable"].split()[0]
    #    if '/' in exec_path_check or '\\' in exec_path_check: # パス区切り文字が含まれているか
    #         if not os.path.exists(exec_path_check):
    #              print(f"Warning: Potential executable path '{exec_path_check}' derived from '{CONFIG['solver_executable']}' might not exist.", file=sys.stderr)

    print(f"\nCreating/Loading Optuna study: '{CONFIG['study_name']}'")
    print(f"Optimization direction: {CONFIG['direction']}")
    print(f"Number of trials: {CONFIG['n_trials']}")

    # 4. Optuna Studyの作成または読み込み
    #    DBに保存して中断・再開したい場合は storage を指定する
    #    例: storage="sqlite:///optuna_study.db"
    study = optuna.create_study(
        study_name=CONFIG["study_name"],
        direction=CONFIG["direction"],
        # storage=None, # In-memory storage (default)
        # load_if_exists=True # DB使用時に既存のStudyを読み込む場合
    )

    # 5. 最適化の実行
    print("\nStarting optimization loop...")
    try:
        study.optimize(
            lambda trial: objective(trial, CONFIG), # objective関数にconfigを渡す
            n_trials=CONFIG["n_trials"],
            timeout=None, # 時間制限 (秒) を設ける場合
            # callbacks=[...] # コールバック関数を指定する場合
        )
    except KeyboardInterrupt:
        print("\nOptimization interrupted by user (KeyboardInterrupt).")
    except Exception as e:
        print(f"\nAn unexpected error occurred during optimization: {e}", file=sys.stderr)
        import traceback
        traceback.print_exc(file=sys.stderr) # スタックトレースを出力

    # 6. 結果の表示
    print("\n----- Optimization Finished -----")
    print(f"Number of finished trials: {len(study.trials)}")

    # 最良試行の結果を表示 (ただし、成功した試行がない場合を除く)
    if study.best_trial:
        print("\nBest trial:")
        print(f"  Value (Score): {study.best_trial.value}")
        print("  Params: ")
        for key, value in study.best_trial.params.items():
            print(f"{key} = {value}")
    elif len(study.trials) > 0:
         print("\nNo successful trials completed that returned a valid score.")
         print("Check error messages in the logs above.")
    else:
         print("\nNo trials were completed.")

    # 他の分析情報 (任意、必要なら matplotlib や plotly が必要)
    # try:
    #    if study.trials:
    #        print("\nGenerating plots (requires plotly)...")
    #        fig_history = optuna.visualization.plot_optimization_history(study)
    #        fig_importance = optuna.visualization.plot_param_importances(study)
    #        fig_history.show()
    #        fig_importance.show()
    # except (ImportError, RuntimeError) as e:
    #    print(f"\nCould not generate plots: {e}. Install plotly and/or kaleido if needed.")
    # except Exception as e:
    #    print(f"\nAn unexpected error occurred during plot generation: {e}", file=sys.stderr)


    print("\nOptuna script finished.")
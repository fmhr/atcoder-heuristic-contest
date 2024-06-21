import subprocess
import optuna

# コマンドを実行してスコアを取得する関数
def run_command(a, b):
    # コマンドを組み立てる
    # cmd = f"fj tests 300 --no-table --cloud --args=\"--ws={a} --cp={b}\""
    cmd = f"fj test --args=\"--ws={a} --cp={b}\"" # 1ケースのみ
    
    # コマンドを実行
    result = subprocess.run(cmd, shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)

    # コマンドのエラー出力を確認
    #if result.stderr:
        #raise ValueError(f"Command error: {result.stderr.decode()}")

    # 結果をパースしてスコアを返す（この部分はコマンドの出力形式に依存します）
    # ここでは、stdoutから直接数値を取得すると仮定しています
    # print(result.stdout.decode())
    # score = float(result.stdout.decode())
    score_str = result.stdout.decode().strip()
    score = float(score_str) if score_str else 0
    if score == 0:
        print(result.stdout.decode())
    return score

# Optunaの目的関数
def objective(trial):
    # 最適化する変数を指定
    ws = trial.suggest_float('ws', 0.01, 10.0)
    cp = trial.suggest_float('cp', 0.01, 10.0)
    # スコアを取得
    score = run_command(ws, cp)
    
    return score

# 最適化のセッションを開始
study = optuna.create_study(direction="maximize")  # minimizeかmaximizeは目的に応じて変更してください
study.optimize(objective, n_trials=1000)  # n_trialsは試行回数。必要に応じて調整してください

# 最適化結果を表示
print(f"Best ws: {study.best_params['ws']}")
print(f"Best cp: {study.best_params['cp']}")
print(f"Best score: {study.best_value}")

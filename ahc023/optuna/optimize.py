import subprocess
import optuna

# コマンドを実行してスコアを取得する関数
def run_command(a):
    # コマンドを組み立てる
    # cmd = f"/Users/fumihiro/fmj/bin/fmj -app=run -end=10 -cmdArgs=\"-timeAddition={a} -maxLength={b}\""
    cmd = f"/Users/fumihiro/fmj/bin/fmj -app=run -end=10 -cmdArgs=\"-maxRegionSize={a}\""
    
    # コマンドを実行
    result = subprocess.run(cmd, shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)

     # コマンドのエラー出力を確認
#    if result.stderr:
#        raise ValueError(f"Command error: {result.stderr.decode()}")

    # 結果をパースしてスコアを返す（この部分はコマンドの出力形式に依存します）
    # ここでは、stdoutから直接数値を取得すると仮定しています
    #print(result.stdout.decode())
    score = float(result.stdout.decode())
    return score

# Optunaの目的関数
def objective(trial):
    # 最適化する変数を指定
    # time_Addition = trial.suggest_int('timeAddition', 0, 5)
    # max_Length = trial.suggest_int('maxLength', 0, 200)
    maxRegionSize = trial.suggest_int('maxRegionSize', 3, 40)
    
    # スコアを取得
    score = run_command(maxRegionSize)
    
    return score

# 最適化のセッションを開始
study = optuna.create_study(direction="maximize")  # minimizeかmaximizeは目的に応じて変更してください
study.optimize(objective, n_trials=200)  # n_trialsは試行回数。必要に応じて調整してください

# 最適化結果を表示
# print(f"Best timeAddition: {study.best_params['timeAddition']}")
# print(f"Best maxLength: {study.best_params['maxLength']}")
print(f"Best maxLength: {study.best_params['maxRegionSize']}")
print(f"Best score: {study.best_value}")

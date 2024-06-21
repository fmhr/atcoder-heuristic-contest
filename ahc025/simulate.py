import numpy as np
import scipy.stats as stats

def simulate_confidence_interval(lambda_value, max_samples, target_accuracy):
    """
    lambda_value: 指数分布のλ値
    max_samples: 試行する最大サンプル数
    target_accuracy: 目標とする信頼区間の精度（平均値の±）
    """
    z_value = 1.96  # 95% 信頼区間のZ値
    current_samples = 10  # 開始サンプル数、これは好みに応じて調整可能

    while current_samples <= max_samples:
        # 指数分布からのサンプリング
        samples = np.random.exponential(scale=1 / lambda_value, size=current_samples)
        # 値の調整
        adjusted_samples = np.maximum(1, np.round(samples))

        # サンプル統計の計算
        sample_mean = np.mean(adjusted_samples)
        sample_std = np.std(adjusted_samples, ddof=1)  # 不偏標準偏差
        std_error = sample_std / np.sqrt(current_samples)

        # 信頼区間の計算
        margin_of_error = z_value * std_error
        confidence_interval = (sample_mean - margin_of_error, sample_mean + margin_of_error)

        # 信頼区間の精度が目標に達しているかチェック
        if margin_of_error <= target_accuracy:
            return current_samples, confidence_interval

        # サンプル数の増加
        current_samples += 10  # 増分は好みに応じて調整可能

    raise ValueError(f"目標精度に達するには、{max_samples}以上のサンプルが必要です。")

# 使用例
lambda_value = 1e-5  # λ = 10^-5
max_samples = 10000  # 最大試行サンプル数（これは十分大きく設定する必要がある）
target_accuracy = 500  # これは例えば、平均値の±500とする

try:
    optimal_sample_size, conf_interval = simulate_confidence_interval(lambda_value, max_samples, target_accuracy)
    print(f"所望の精度に達するためのサンプルサイズ: {optimal_sample_size}")
    print(f"95％信頼区間: {conf_interval}")
except ValueError as e:
    print(e)

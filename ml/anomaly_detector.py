#!/usr/bin/env python3
"""
Anomaly Detection Service for CloudBridge Relay

This service provides ML-based anomaly detection for:
- Network traffic anomalies
- Security threats
- Performance anomalies
- User behavior anomalies
"""

import numpy as np
import pandas as pd
from sklearn.ensemble import IsolationForest
from sklearn.preprocessing import StandardScaler
from sklearn.neural_network import MLPRegressor
import joblib
import logging
from typing import Dict, List, Tuple, Optional
import json
import time
from datetime import datetime, timedelta
import redis
import requests
import structlog
import os
from dotenv import load_dotenv

# Configure logging
structlog.configure(
    processors=[
        structlog.processors.TimeStamper(fmt="iso"),
        structlog.processors.JSONRenderer()
    ]
)
logger = structlog.get_logger(__name__)

load_dotenv(dotenv_path=os.path.join(os.path.dirname(__file__), '.env'), override=True)


class AnomalyDetector:
    """Main anomaly detection service"""
    
    def __init__(self, redis_host: str = None, redis_port: int = None):
        redis_host = redis_host or os.getenv('REDIS_HOST', 'localhost')
        redis_port = int(redis_port or os.getenv('REDIS_PORT', 6379))
        iforest_contamination = float(os.getenv('IFOREST_CONTAMINATION', 0.1))
        iforest_n_estimators = int(os.getenv('IFOREST_N_ESTIMATORS', 100))
        mlp_hidden_layers = tuple(int(x) for x in os.getenv('MLP_HIDDEN_LAYERS', '100,50').split(','))
        mlp_max_iter = int(os.getenv('MLP_MAX_ITER', 1000))
        self.redis_client = redis.Redis(host=redis_host, port=redis_port, decode_responses=True)
        self.scaler = StandardScaler()
        self.isolation_forest = IsolationForest(
            contamination=iforest_contamination,
            random_state=42,
            n_estimators=iforest_n_estimators
        )
        self.performance_predictor = MLPRegressor(
            hidden_layer_sizes=mlp_hidden_layers,
            max_iter=mlp_max_iter,
            random_state=42
        )
        self.models_trained = False
        
    def extract_features(self, data: Dict) -> np.ndarray:
        """Extract features from raw data"""
        features = []
        
        # Network features
        features.extend([
            data.get('bytes_sent', 0),
            data.get('bytes_received', 0),
            data.get('packets_sent', 0),
            data.get('packets_received', 0),
            data.get('connection_duration', 0),
            data.get('requests_per_second', 0),
            data.get('error_rate', 0),
        ])
        
        # Time-based features
        timestamp = datetime.fromisoformat(data.get('timestamp', datetime.now().isoformat()))
        features.extend([
            timestamp.hour,
            timestamp.weekday(),
            timestamp.month,
        ])
        
        # User behavior features
        features.extend([
            data.get('user_id_hash', 0) % 1000,  # Hash for privacy
            data.get('session_duration', 0),
            data.get('requests_in_session', 0),
            data.get('unique_endpoints', 0),
        ])
        
        # Security features
        features.extend([
            data.get('failed_auth_attempts', 0),
            data.get('suspicious_ips', 0),
            data.get('rate_limit_violations', 0),
            data.get('ssl_errors', 0),
        ])
        
        return np.array(features).reshape(1, -1)
    
    def train_models(self, historical_data: List[Dict]):
        """Train anomaly detection models"""
        logger.info("Training anomaly detection models...")
        
        # Extract features from historical data
        features_list = []
        performance_targets = []
        
        for data_point in historical_data:
            features = self.extract_features(data_point)
            features_list.append(features.flatten())
            
            # For performance prediction
            if 'latency' in data_point:
                performance_targets.append(data_point['latency'])
        
        if len(features_list) < 100:
            logger.warning("Insufficient data for training. Need at least 100 samples.", sample_count=len(features_list))
            return False
        
        # Convert to numpy arrays
        X = np.array(features_list)
        y_performance = np.array(performance_targets) if performance_targets else None
        
        # Scale features
        X_scaled = self.scaler.fit_transform(X)
        
        # Train isolation forest for anomaly detection
        self.isolation_forest.fit(X_scaled)
        
        # Train performance predictor if we have targets
        if y_performance is not None and len(y_performance) > 0:
            self.performance_predictor.fit(X_scaled, y_performance)
        
        self.models_trained = True
        logger.info("Models trained successfully!", sample_count=len(features_list))
        return True
    
    def detect_anomalies(self, data: Dict) -> Dict:
        """Detect anomalies in real-time data"""
        if not self.models_trained:
            logger.warning("Detection called before models trained.")
            return {
                'anomaly_score': 0.0,
                'is_anomaly': False,
                'confidence': 0.0,
                'anomaly_type': 'unknown',
                'recommendations': ['Models not trained yet']
            }
        
        try:
            # Extract features
            features = self.extract_features(data)
            logger.debug("Extracted features", features=features.tolist())
            features_scaled = self.scaler.transform(features)
            
            # Get anomaly score from isolation forest
            anomaly_score = self.isolation_forest.decision_function(features_scaled)[0]
            
            # Convert to probability (0-1 scale)
            anomaly_probability = 1 - np.exp(anomaly_score)
            
            # Determine if it's an anomaly (threshold-based)
            is_anomaly = anomaly_probability > 0.7
            
            # Classify anomaly type
            anomaly_type = self._classify_anomaly_type(data, anomaly_probability)
            
            # Generate recommendations
            recommendations = self._generate_recommendations(data, anomaly_type, anomaly_probability)
            logger.info("Anomaly detection result", anomaly_score=anomaly_probability, is_anomaly=is_anomaly, anomaly_type=anomaly_type, recommendations=recommendations)
            
            return {
                'anomaly_score': float(anomaly_probability),
                'is_anomaly': bool(is_anomaly),
                'confidence': float(min(anomaly_probability * 1.2, 1.0)),
                'anomaly_type': anomaly_type,
                'recommendations': recommendations,
                'timestamp': datetime.now().isoformat(),
                'features_used': len(features.flatten())
            }
            
        except Exception as e:
            logger.error("Error in anomaly detection", error=str(e), data=data)
            return {
                'anomaly_score': 0.0,
                'is_anomaly': False,
                'confidence': 0.0,
                'anomaly_type': 'error',
                'recommendations': [f'Detection error: {str(e)}'],
                'timestamp': datetime.now().isoformat()
            }
    
    def predict_performance(self, data: Dict) -> Dict:
        """Predict performance metrics"""
        if not self.models_trained:
            return {
                'predicted_latency': 0.0,
                'confidence': 0.0,
                'recommendations': ['Performance model not trained yet']
            }
        
        try:
            features = self.extract_features(data)
            features_scaled = self.scaler.transform(features)
            
            predicted_latency = self.performance_predictor.predict(features_scaled)[0]
            
            return {
                'predicted_latency': float(predicted_latency),
                'confidence': 0.8,  # Placeholder
                'recommendations': self._generate_performance_recommendations(data, predicted_latency)
            }
            
        except Exception as e:
            logger.error(f"Error in performance prediction: {e}")
            return {
                'predicted_latency': 0.0,
                'confidence': 0.0,
                'recommendations': [f'Prediction error: {str(e)}']
            }
    
    def _classify_anomaly_type(self, data: Dict, anomaly_score: float) -> str:
        """Classify the type of anomaly"""
        if anomaly_score > 0.9:
            return 'critical_security'
        elif anomaly_score > 0.8:
            return 'security_threat'
        elif anomaly_score > 0.7:
            return 'performance_issue'
        elif anomaly_score > 0.6:
            return 'behavioral_anomaly'
        else:
            return 'normal'
    
    def _generate_recommendations(self, data: Dict, anomaly_type: str, score: float) -> List[str]:
        """Generate recommendations based on anomaly type"""
        recommendations = []
        
        if anomaly_type == 'critical_security':
            recommendations.extend([
                'Immediate IP blocking recommended',
                'Enable enhanced monitoring',
                'Review authentication logs',
                'Consider rate limiting'
            ])
        elif anomaly_type == 'security_threat':
            recommendations.extend([
                'Monitor user activity closely',
                'Check for credential stuffing',
                'Review recent login attempts',
                'Consider MFA enforcement'
            ])
        elif anomaly_type == 'performance_issue':
            recommendations.extend([
                'Check server resources',
                'Review recent deployments',
                'Monitor database performance',
                'Consider scaling up'
            ])
        elif anomaly_type == 'behavioral_anomaly':
            recommendations.extend([
                'Monitor user patterns',
                'Check for account compromise',
                'Review access patterns',
                'Consider additional verification'
            ])
        
        return recommendations
    
    def _generate_performance_recommendations(self, data: Dict, predicted_latency: float) -> List[str]:
        """Generate performance recommendations"""
        recommendations = []
        
        if predicted_latency > 1000:  # >1 second
            recommendations.extend([
                'High latency predicted - consider optimization',
                'Check database query performance',
                'Review caching strategy',
                'Consider CDN implementation'
            ])
        elif predicted_latency > 500:  # >500ms
            recommendations.extend([
                'Moderate latency - monitor closely',
                'Check resource utilization',
                'Review connection pooling'
            ])
        
        return recommendations
    
    def save_models(self, filepath: str):
        """Save trained models to disk"""
        if self.models_trained:
            try:
                joblib.dump({
                    'scaler': self.scaler,
                    'isolation_forest': self.isolation_forest,
                    'performance_predictor': self.performance_predictor
                }, filepath)
                logger.info("Models saved", filepath=filepath)
            except Exception as e:
                logger.error("Error saving models", error=str(e), filepath=filepath)
    
    def load_models(self, filepath: str):
        """Load trained models from disk"""
        try:
            models = joblib.load(filepath)
            self.scaler = models['scaler']
            self.isolation_forest = models['isolation_forest']
            self.performance_predictor = models['performance_predictor']
            self.models_trained = True
            logger.info("Models loaded from file", filepath=filepath)
        except Exception as e:
            logger.error("Error loading models", error=str(e), filepath=filepath)


class MLService:
    """Main ML service for CloudBridge Relay"""
    
    def __init__(self, host: str = '0.0.0.0', port: int = 8082):
        self.host = host
        self.port = port
        self.detector = AnomalyDetector()
        self.metrics = {
            'requests_processed': 0,
            'anomalies_detected': 0,
            'avg_processing_time': 0.0
        }
    
    def process_request(self, data: Dict) -> Dict:
        """Process incoming request for anomaly detection"""
        start_time = time.time()
        
        try:
            # Detect anomalies
            anomaly_result = self.detector.detect_anomalies(data)
            
            # Predict performance
            performance_result = self.detector.predict_performance(data)
            
            # Update metrics
            processing_time = time.time() - start_time
            self.metrics['requests_processed'] += 1
            if anomaly_result['is_anomaly']:
                self.metrics['anomalies_detected'] += 1
            
            # Update average processing time
            current_avg = self.metrics['avg_processing_time']
            total_requests = self.metrics['requests_processed']
            self.metrics['avg_processing_time'] = (
                (current_avg * (total_requests - 1) + processing_time) / total_requests
            )
            
            return {
                'success': True,
                'anomaly_detection': anomaly_result,
                'performance_prediction': performance_result,
                'processing_time_ms': round(processing_time * 1000, 2),
                'timestamp': datetime.now().isoformat()
            }
            
        except Exception as e:
            logger.error(f"Error processing request: {e}")
            return {
                'success': False,
                'error': str(e),
                'timestamp': datetime.now().isoformat()
            }
    
    def get_metrics(self) -> Dict:
        """Get service metrics"""
        return {
            **self.metrics,
            'models_trained': self.detector.models_trained,
            'timestamp': datetime.now().isoformat()
        }


# Example usage
if __name__ == "__main__":
    # Initialize ML service
    ml_service = MLService()
    
    # Example data point
    sample_data = {
        'timestamp': datetime.now().isoformat(),
        'bytes_sent': 1024,
        'bytes_received': 2048,
        'packets_sent': 10,
        'packets_received': 20,
        'connection_duration': 30,
        'requests_per_second': 5,
        'error_rate': 0.01,
        'user_id_hash': 12345,
        'session_duration': 300,
        'requests_in_session': 50,
        'unique_endpoints': 5,
        'failed_auth_attempts': 0,
        'suspicious_ips': 0,
        'rate_limit_violations': 0,
        'ssl_errors': 0
    }
    
    # Process sample data
    result = ml_service.process_request(sample_data)
    print(json.dumps(result, indent=2))
    
    # Get metrics
    metrics = ml_service.get_metrics()
    print(json.dumps(metrics, indent=2)) 